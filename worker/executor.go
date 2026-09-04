package worker

import (
	"context"
	"devcloud/ent"
	"devcloud/ent/resource"
	"devcloud/ent/task"
	"devcloud/provider"
	"devcloud/repo"
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

func (e *Executor) execute(ctx context.Context, t *ent.Task) error {
	switch t.Type {
	case task.TypeCREATE_RESOURCE:
		return e.createResource(ctx, t)

	case task.TypeSTART_RESOURCE:
		return e.startResource(ctx, t)

	case task.TypeSTOP_RESOURCE:
		return e.stopResource(ctx, t)

	case task.TypeRESTART_RESOURCE:
		return e.restartResource(ctx, t)

	case task.TypeDELETE_RESOURCE:
		return e.deleteResource(ctx, t)

	default:
		return fmt.Errorf("unsupported task type: %s", t.Type)
	}
}

func (e *Executor) createResource(ctx context.Context, t *ent.Task) error {
	rsc, err := e.repo.Resource.Get(ctx, t.ResourceID)
	if err != nil {
		return err
	}

	p, err := e.registry.Get(rsc.Provider)
	if err != nil {
		return err
	}

	runtimeID, err := p.Create(ctx, &provider.ResourceSpec{
		Name:       rsc.Name,
		Type:       rsc.Type,
		Config:     rsc.Config,
		Credential: rsc.Credential,
	})
	if err != nil {
		return err
	}

	return e.repo.Resource.MarkCreated(ctx, rsc.ID, runtimeID)
}

func (e *Executor) startResource(ctx context.Context, t *ent.Task) error {
	rsc, err := e.repo.Resource.Get(ctx, t.ResourceID)
	if err != nil {
		return err
	}

	p, err := e.registry.Get(rsc.Provider)
	if err != nil {
		return err
	}

	if err := e.repo.Resource.MarkStarting(ctx, rsc.ID); err != nil {
		return err
	}

	if err := p.Start(ctx, rsc.RuntimeID); err != nil {
		_ = e.repo.Resource.MarkFailed(ctx, rsc.ID, resource.StatusSTARTING)
		return err
	}

	return e.repo.Resource.MarkRunning(ctx, rsc.ID, resource.StatusSTARTING)
}

func (e *Executor) restartResource(ctx context.Context, t *ent.Task) error {
	rsc, err := e.repo.Resource.Get(ctx, t.ResourceID)
	if err != nil {
		return err
	}

	p, err := e.registry.Get(rsc.Provider)
	if err != nil {
		return err
	}

	if err := e.repo.Resource.MarkRestarting(ctx, rsc.ID); err != nil {
		return err
	}

	if err := p.Restart(ctx, rsc.RuntimeID); err != nil {
		_ = e.repo.Resource.MarkFailed(ctx, rsc.ID, resource.StatusRESTARTING)
		return err
	}

	return e.repo.Resource.MarkRunning(ctx, rsc.ID, resource.StatusRESTARTING)
}

func (e *Executor) stopResource(ctx context.Context, t *ent.Task) error {
	rsc, err := e.repo.Resource.Get(ctx, t.ResourceID)
	if err != nil {
		return err
	}

	p, err := e.registry.Get(rsc.Provider)
	if err != nil {
		return err
	}

	if err := e.repo.Resource.MarkStopping(ctx, rsc.ID); err != nil {
		return err
	}

	if err := p.Stop(ctx, rsc.RuntimeID); err != nil {
		_ = e.repo.Resource.MarkFailed(ctx, rsc.ID, resource.StatusSTOPPING)
		return err
	}

	return e.repo.Resource.MarkStopped(ctx, rsc.ID)
}

func (e *Executor) deleteResource(ctx context.Context, t *ent.Task) error {
	rsc, err := e.repo.Resource.Get(ctx, t.ResourceID)
	if err != nil {
		return err
	}

	p, err := e.registry.Get(rsc.Provider)
	if err != nil {
		return err
	}

	if err := e.repo.Resource.MarkDeleting(ctx, rsc.ID); err != nil {
		return err
	}

	if err := p.Delete(ctx, rsc.RuntimeID); err != nil {
		_ = e.repo.Resource.MarkFailed(ctx, rsc.ID, resource.StatusDELETING)
		return err
	}

	return e.repo.Resource.MarkDeleted(ctx, rsc.ID)
}
