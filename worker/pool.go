package worker

import (
	"context"
	"devcloud/provider"
	"devcloud/repo"
	"log/slog"
	"sync"
	"time"
)

type Pool struct {
	ctx    context.Context
	cancel context.CancelFunc

	wg      sync.WaitGroup
	workers int

	executor    *Executor
	taskManager *TaskManager
}

func NewPool(
	pCtx context.Context,
	workers int,
	repo *repo.Repo,
	registry *provider.Registry,
) *Pool {
	if workers < 0 {
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

func (p *Pool) Start() {
	for range p.workers {
		p.wg.Go(func() {
			newWorker(
				p.taskManager,
				p.executor,
			).run(p.ctx)
		})
	}
}

func (p *Pool) Stop() {
	p.cancel()
	p.wg.Wait()
}

type worker struct {
	taskManager *TaskManager
	executor    *Executor
}

func newWorker(
	taskManager *TaskManager,
	executor *Executor,
) *worker {
	return &worker{
		taskManager: taskManager,
		executor:    executor,
	}
}

func (w *worker) run(ctx context.Context) {
	d := time.Second
	timer := time.NewTimer(d)
	for {
		if ctx.Err() != nil {
			return
		}

		if err := w.process(ctx); err != nil {
			slog.Error("worker execute")
			// 避免频繁报错导致 CPU 空转
			select {
			case <-ctx.Done():
				return
			case <-timer.C:
				timer.Reset(d)
			}
		}
	}
}

func (w *worker) process(ctx context.Context) error {
	task, err := w.taskManager.fetch(ctx)
	if err != nil {
		return err
	}

	if err := w.executor.execute(ctx, task); err != nil {
		return w.taskManager.fail(ctx, task.ID, err.Error())
	}
	return w.taskManager.success(ctx, task.ID)
}
