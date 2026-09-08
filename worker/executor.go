package worker

import (
	"context"
	"devcloud/ent"
	"devcloud/ent/task"
	"devcloud/provider"
	"devcloud/repo"
	"devcloud/repo/repos"
	"devcloud/transition"
	"fmt"
)

type Executor struct {
	repo     *repo.Repo
	registry *provider.Registry
}

func newExecutor(
	repo *repo.Repo,
	registry *provider.Registry,
) *Executor {
	return &Executor{
		repo:     repo,
		registry: registry,
	}
}

// TODO：错误重试策略
func (e *Executor) execute(ctx context.Context, t *repos.ClaimTaskResult) error {
	r, err := e.repo.Resource.Get(ctx, t.ResourceID)
	if err != nil {
		return err
	}

	switch t.Type {
	case task.TypeCREATE_RESOURCE:
		return e.createResource(ctx, r)

	case task.TypeSTART_RESOURCE:
		return e.startResource(ctx, r)

	case task.TypeSTOP_RESOURCE:
		return e.stopResource(ctx, r)

	case task.TypeRESTART_RESOURCE:
		return e.restartResource(ctx, r)

	case task.TypeREMOVE_RESOURCE:
		return e.removeResource(ctx, r)

	default:
		return fmt.Errorf("unsupported task type: %s", t.Type)
	}
}

func (e *Executor) runOperation(
	ctx context.Context,
	r *ent.Resource,
	toMid transition.ResourceEvent,
	toFinish transition.ResourceEvent,
	op func(ctx context.Context, runtimeID string) error,
) error {
	mid, err := transition.Resource(r.Status, toMid)
	if err != nil {
		return err
	}

	fin, err := transition.Resource(mid, toFinish)
	if err != nil {
		return err
	}

	if err := e.repo.Resource.SetStatus(ctx, r.ID, mid); err != nil {
		return err
	}

	if err := op(ctx, r.RuntimeID); err != nil {
		_ = e.repo.Resource.SetStatusError(ctx, r.ID, r.Status, err.Error())
		return err
	}

	return e.repo.Resource.SetStatus(ctx, r.ID, fin)
}

func (e *Executor) createResource(ctx context.Context, r *ent.Resource) error {
	p, err := e.registry.Get(r.Provider)
	if err != nil {
		return err
	}

	if _, err := transition.Resource(r.Status, transition.ResourceCreated); err != nil {
		return err
	}

	runtimeID, err := p.Create(ctx, provider.ResourceSpec{
		Name:       r.Name,
		Image:      r.Image,
		Config:     r.Config,
		Credential: provider.Credential{}, // TODO：r.Credential 解密后填入
	})
	if err != nil {
		failed, terr := transition.Resource(r.Status, transition.ResourceCreateFailed)
		if terr != nil {
			failed = r.Status
		}
		_ = e.repo.Resource.SetStatusError(ctx, r.ID, failed, err.Error())
		return err
	}

	return e.repo.Resource.SetStatusRUNNING(
		ctx,
		r.ID,
		runtimeID,
	)
}

func (e *Executor) startResource(ctx context.Context, r *ent.Resource) error {
	p, err := e.registry.Get(r.Provider)
	if err != nil {
		return err
	}

	return e.runOperation(ctx, r, transition.ResourceStart, transition.ResourceStarted, p.Start)
}

func (e *Executor) restartResource(ctx context.Context, r *ent.Resource) error {
	p, err := e.registry.Get(r.Provider)
	if err != nil {
		return err
	}

	return e.runOperation(ctx, r, transition.ResourceRestart, transition.ResourceRestarted, p.Restart)
}

func (e *Executor) stopResource(ctx context.Context, r *ent.Resource) error {
	p, err := e.registry.Get(r.Provider)
	if err != nil {
		return err
	}

	return e.runOperation(ctx, r, transition.ResourceStop, transition.ResourceStopped, p.Stop)
}

func (e *Executor) removeResource(ctx context.Context, r *ent.Resource) error {
	p, err := e.registry.Get(r.Provider)
	if err != nil {
		return err
	}

	return e.runOperation(ctx, r, transition.ResourceRemove, transition.ResourceRemoved, p.Remove)
}
