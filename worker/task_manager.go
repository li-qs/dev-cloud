package worker

import (
	"context"
	"devcloud/ent"
	"devcloud/repo"
	"math/rand"
	"time"
)

const (
	baseBackoff = time.Second
	maxBackoff  = 30 * time.Second
)

type TaskManager struct {
	repo *repo.Repo
}

func newTaskManager(repo *repo.Repo) *TaskManager {
	return &TaskManager{repo: repo}
}

// fetch 领取一条到期任务并建立租约；无任务时返回 apperr.ErrNoPendingTask。
func (t *TaskManager) fetch(ctx context.Context, workerID string) (*ent.Task, error) {
	return t.repo.Task.ClaimPendingTask(ctx, workerID)
}

// success 将任务标记为成功（受 workerID 栅栏保护）。
func (t *TaskManager) success(ctx context.Context, id int, workerID string) error {
	return t.repo.Task.SetSuccess(ctx, id, workerID)
}

// heartbeat 刷新任务心跳，维持租约。
func (t *TaskManager) heartbeat(ctx context.Context, id int, workerID string) error {
	return t.repo.Task.Heartbeat(ctx, id, workerID)
}

// fail 根据已重试次数决定是重新入队（指数退避）还是终结为 FAILED。
func (t *TaskManager) fail(ctx context.Context, task *ent.Task, workerID string, errMsg string) error {
	if task.Attempts < task.MaxAttempts {
		return t.repo.Task.Retry(
			ctx,
			task.ID,
			workerID,
			errMsg,
			time.Now().Add(retryBackoff(task.Attempts)),
		)
	}
	return t.repo.Task.SetFailed(ctx, task.ID, workerID, errMsg)
}

// retryBackoff 返回尝试次数对应的退避时长，带抖动避免多个任务同时重试。
func retryBackoff(attempts int) time.Duration {
	if attempts < 1 {
		attempts = 1
	}
	shift := attempts - 1
	if shift > 5 {
		shift = 5
	}

	d := baseBackoff << shift
	if d > maxBackoff {
		d = maxBackoff
	}

	// 抖动：50% ~ 100%
	return d/2 + time.Duration(rand.Int63n(int64(d/2)+1))
}

// recover 回收心跳超时、仍处于 RUNNING 的任务。
func (t *TaskManager) recover(ctx context.Context, timeout time.Duration, limit int) (int, error) {
	return t.repo.Task.RecoverStuck(ctx, timeout, limit)
}
