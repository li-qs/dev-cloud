package worker

import (
	"context"
	"devcloud/apperr"
	"devcloud/provider"
	"devcloud/repo"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"sync"
	"time"
)

const (
	// 空闲轮询退避范围。
	minIdle = 200 * time.Millisecond
	maxIdle = 2 * time.Second

	// 任务租约：每 heartbeatInterval 刷新一次心跳，超过 heartbeatTimeout 视为持有者失效。
	heartbeatInterval = 5 * time.Second
	heartbeatTimeout  = 30 * time.Second
	// 恢复扫描的周期与单次上限。
	recoveryInterval = 10 * time.Second
	recoveryLimit    = 100
)

type Pool struct {
	ctx    context.Context
	cancel context.CancelFunc

	wg      sync.WaitGroup
	workers int

	executor    *Executor
	taskManager *TaskManager
}

// NewPool 创建 worker 池，workers 至少为 1。
func NewPool(
	pCtx context.Context,
	workers int,
	repo *repo.Repo,
	registry *provider.Registry,
) *Pool {
	if workers <= 0 {
		workers = 1
	}

	ctx, cancel := context.WithCancel(pCtx)

	return &Pool{
		ctx:         ctx,
		cancel:      cancel,
		workers:     workers,
		taskManager: newTaskManager(repo),
		executor:    newExecutor(repo, registry),
	}
}

// Start 启动一个恢复协程和一组 worker；每个 worker 用唯一 id 作为租约持有者标识。
func (p *Pool) Start() {
	p.wg.Go(p.recoverLoop)

	host, err := os.Hostname()
	if err != nil {
		host = "worker"
	}

	for i := range p.workers {
		workerID := fmt.Sprintf("%s-%d", host, i)
		p.wg.Go(func() {
			newWorker(
				workerID,
				p.taskManager,
				p.executor,
			).run(p.ctx)
		})
	}
}

// Stop 取消上下文并等待所有协程退出（优雅关闭）。
func (p *Pool) Stop() {
	p.cancel()
	p.wg.Wait()
}

// recoverLoop 定期回收心跳超时、仍卡在 RUNNING 的任务（worker 崩溃/卡死兜底）。
func (p *Pool) recoverLoop() {
	ticker := time.NewTicker(recoveryInterval)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			n, err := p.taskManager.recover(p.ctx, heartbeatTimeout, recoveryLimit)
			if err != nil {
				slog.Error("recover stuck tasks", "error", err)
				continue
			}
			if n > 0 {
				slog.Warn("recovered stuck tasks", "count", n)
			}
		}
	}
}

type worker struct {
	id          string
	taskManager *TaskManager
	executor    *Executor
}

func newWorker(
	id string,
	taskManager *TaskManager,
	executor *Executor,
) *worker {
	return &worker{
		id:          id,
		taskManager: taskManager,
		executor:    executor,
	}
}

// run worker 主循环：有任务就连续处理；空闲或出错时按指数退避轮询。
func (w *worker) run(ctx context.Context) {
	idle := minIdle
	for ctx.Err() == nil {
		err := w.process(ctx)

		switch {
		case err == nil:
			// 有任务被执行，可能还有积压，立即继续。
			idle = minIdle
			continue
		case errors.Is(err, apperr.ErrNoPendingTask):
			// 队列为空，正常退避。
		default:
			slog.Error("worker process", "error", err)
		}

		if !sleep(ctx, jitter(idle)) {
			return
		}
		if idle < maxIdle {
			idle *= 2
			if idle > maxIdle {
				idle = maxIdle
			}
		}
	}
}

// process 领取并执行一条任务：执行期间起协程持续刷新心跳；
// 成功或失败后按 workerID 栅栏更新任务终态/重试。
func (w *worker) process(ctx context.Context) error {
	task, err := w.taskManager.fetch(ctx, w.id)
	if err != nil {
		return err
	}

	// 执行期间持续发送心跳，维持租约。
	execCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	go w.heartbeat(execCtx, task.ID)

	if err := w.executor.execute(ctx, task); err != nil {
		cancel()
		// 服务关闭/超时导致的取消不算业务失败；保留租约，交由心跳超时恢复处理。
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil
		}
		return w.taskManager.fail(ctx, task, w.id, err.Error())
	}

	cancel()
	return w.taskManager.success(ctx, task.ID, w.id)
}

// heartbeat 周期性刷新任务心跳，任务执行结束（ctx 取消）后停止。
func (w *worker) heartbeat(ctx context.Context, taskID int) {
	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := w.taskManager.heartbeat(ctx, taskID, w.id); err != nil {
				slog.Warn("task heartbeat", "task_id", taskID, "worker_id", w.id, "error", err)
			}
		}
	}
}

// sleep 等待 d，期间可被 ctx 取消；返回 false 表示应退出。
func sleep(ctx context.Context, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

// jitter 返回 [d/2, d] 的随机值，避免多个 worker 同时轮询。
func jitter(d time.Duration) time.Duration {
	if d <= 0 {
		return d
	}
	return d/2 + time.Duration(rand.Int63n(int64(d/2)+1))
}
