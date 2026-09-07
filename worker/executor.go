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

func (e *Executor) createResource(ctx context.Context, r *ent.Resource) error {
	p, err := e.registry.Get(r.Provider)
	if err != nil {
		return err
	}

	_, err = transition.Resource(r.Status, transition.ResourceCreated)
	if err != nil {
		return err
	}

	runtimeID, err := p.Create(ctx, provider.ResourceSpec{
		Name:       r.Name,
		Image:      r.Image,
		Config:     r.Config,
		Credential: provider.Credential{}, // TODO：r.Credential 解密后填入
	})
	if err != nil {
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

	nextStatus, err := transition.Resource(r.Status, transition.ResourceStart)
	if err != nil {
		return err
	}

	if err := e.repo.Resource.SetStatus(
		ctx,
		r.ID,
		nextStatus,
	); err != nil {
		return err
	}

	nextStatus, err = transition.Resource(r.Status, transition.ResourceStarted)
	if err != nil {
		return err
	}

	if err := p.Start(ctx, r.RuntimeID); err != nil {
		return err
	}

	return e.repo.Resource.SetStatus(ctx, r.ID, nextStatus)
}

func (e *Executor) restartResource(ctx context.Context, r *ent.Resource) error {
	p, err := e.registry.Get(r.Provider)
	if err != nil {
		return err
	}

	nextStatus, err := transition.Resource(r.Status, transition.ResourceRestart)
	if err != nil {
		return err
	}

	if err := e.repo.Resource.SetStatus(
		ctx,
		r.ID,
		nextStatus,
	); err != nil {
		return err
	}

	nextStatus, err = transition.Resource(r.Status, transition.ResourceRestarted)
	if err != nil {
		return err
	}

	if err := p.Restart(ctx, r.RuntimeID); err != nil {
		return err
	}

	return e.repo.Resource.SetStatus(ctx, r.ID, nextStatus)
}

func (e *Executor) stopResource(ctx context.Context, r *ent.Resource) error {
	p, err := e.registry.Get(r.Provider)
	if err != nil {
		return err
	}

	nextStatus, err := transition.Resource(r.Status, transition.ResourceStop)
	if err != nil {
		return err
	}

	if err := e.repo.Resource.SetStatus(
		ctx,
		r.ID,
		nextStatus,
	); err != nil {
		return err
	}

	nextStatus, err = transition.Resource(r.Status, transition.ResourceStopped)
	if err != nil {
		return err
	}

	if err := p.Stop(ctx, r.RuntimeID); err != nil {
		return err
	}

	return e.repo.Resource.SetStatus(ctx, r.ID, nextStatus)
}

func (e *Executor) removeResource(ctx context.Context, r *ent.Resource) error {
	p, err := e.registry.Get(r.Provider)
	if err != nil {
		return err
	}

	nextStatus, err := transition.Resource(r.Status, transition.ResourceRemove)
	if err != nil {
		return err
	}

	if err := e.repo.Resource.SetStatus(
		ctx,
		r.ID,
		nextStatus,
	); err != nil {
		return err
	}

	nextStatus, err = transition.Resource(r.Status, transition.ResourceRemoved)
	if err != nil {
		return err
	}

	if err := p.Remove(ctx, r.RuntimeID); err != nil {
		return err
	}

	return e.repo.Resource.SetStatus(ctx, r.ID, nextStatus)
}
