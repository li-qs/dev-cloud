package worker

import (
	"context"
	"time"
)

type Worker struct {
	taskFetcher *TaskFetcher
	executor    *Executor
}

func newWorker(
	taskFetcher *TaskFetcher,
	executor *Executor,
) *Worker {
	return &Worker{
		taskFetcher: taskFetcher,
		executor:    executor,
	}
}

func (w *Worker) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if err := w.process(ctx); err != nil {
			// TODO：处理错误
		}
	}
}

func (w *Worker) process(ctx context.Context) error {
	task, err := w.taskFetcher.fetch(ctx)
	if err != nil {
		return err
	}

	taskCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	return w.executor.execute(taskCtx, task)
}
