package worker

import (
	"context"
	"devcloud/ent"
	"devcloud/ent/resource"
	"devcloud/ent/task"
	"devcloud/provider"
	"devcloud/repo"
	"devcloud/repo/repos"
	"fmt"
)

type Executor struct {
	repo     *repo.Repo
	registry *provider.Registry
}

func NewExecutor(
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

	case task.TypeDELETE_RESOURCE:
		return e.deleteResource(ctx, r)

	default:
		return fmt.Errorf("unsupported task type: %s", t.Type)
	}
}

func (e *Executor) createResource(ctx context.Context, r *ent.Resource) error {
	p, err := e.registry.Get(r.Provider)
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

	return e.repo.Resource.MarkCreated(ctx, r.ID, runtimeID)
}

func (e *Executor) startResource(ctx context.Context, r *ent.Resource) error {
	p, err := e.registry.Get(r.Provider)
	if err != nil {
		return err
	}

	if err := e.repo.Resource.MarkStarting(ctx, r.ID); err != nil {
		return err
	}

	if err := p.Start(ctx, r.RuntimeID); err != nil {
		_ = e.repo.Resource.MarkFailed(ctx, r.ID, resource.StatusSTARTING)
		return err
	}

	return e.repo.Resource.MarkRunning(ctx, r.ID, resource.StatusSTARTING)
}

func (e *Executor) restartResource(ctx context.Context, r *ent.Resource) error {
	p, err := e.registry.Get(r.Provider)
	if err != nil {
		return err
	}

	if err := e.repo.Resource.MarkRestarting(ctx, r.ID); err != nil {
		return err
	}

	if err := p.Restart(ctx, r.RuntimeID); err != nil {
		_ = e.repo.Resource.MarkFailed(ctx, r.ID, resource.StatusRESTARTING)
		return err
	}

	return e.repo.Resource.MarkRunning(ctx, r.ID, resource.StatusRESTARTING)
}

func (e *Executor) stopResource(ctx context.Context, r *ent.Resource) error {
	p, err := e.registry.Get(r.Provider)
	if err != nil {
		return err
	}

	if err := e.repo.Resource.MarkStopping(ctx, r.ID); err != nil {
		return err
	}

	if err := p.Stop(ctx, r.RuntimeID); err != nil {
		_ = e.repo.Resource.MarkFailed(ctx, r.ID, resource.StatusSTOPPING)
		return err
	}

	return e.repo.Resource.MarkStopped(ctx, r.ID)
}

func (e *Executor) deleteResource(ctx context.Context, r *ent.Resource) error {
	p, err := e.registry.Get(r.Provider)
	if err != nil {
		return err
	}

	if err := e.repo.Resource.MarkDeleting(ctx, r.ID); err != nil {
		return err
	}

	if err := p.Remove(ctx, r.RuntimeID); err != nil {
		_ = e.repo.Resource.MarkFailed(ctx, r.ID, resource.StatusDELETING)
		return err
	}

	return e.repo.Resource.MarkDeleted(ctx, r.ID)
}
