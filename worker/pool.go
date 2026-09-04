package worker

import (
	"context"
	"devcloud/provider"
	"devcloud/repo"
	"sync"
)

type Pool struct {
	ctx    context.Context
	cancel context.CancelFunc

	wg          sync.WaitGroup
	workers     int
	executor    *Executor
	taskFetcher *TaskFetcher
}

func NewPool(
	pCtx context.Context,
	workers int,
	repo *repo.Repo,
	registry *provider.Registry,
) *Pool {
	ctx, cancel := context.WithCancel(pCtx)
	taskFetcher := NewTaskFetcker(repo)
	executor := NewExecutor(repo, registry)
	return &Pool{
		ctx:         ctx,
		cancel:      cancel,
		workers:     workers,
		taskFetcher: taskFetcher,
		executor:    executor,
	}
}

func (p *Pool) Start() {
	// TODO: 处理死掉的 worker，并创建新 worker，保证健壮。
	for range p.workers {
		p.wg.Go(func() {
			worker := newWorker(p.taskFetcher, p.executor)
			worker.run(p.ctx)
		})
	}
}

func (p *Pool) Stop() {
	p.cancel()
	p.wg.Wait()
}
