package service

import (
	"context"
	"devcloud/ent"
	"devcloud/ent/resource"
	"devcloud/ent/task"
	"devcloud/repo"
)

type Resource struct {
	resourceRepo *repo.Resource
	taskRepo     *repo.Task
}

func NewResource(
	resourceRepo *repo.Resource,
	taskRepo *repo.Task,
) *Resource {
	return &Resource{
		resourceRepo: resourceRepo,
		taskRepo:     taskRepo,
	}
}

func (r *Resource) List(ctx context.Context, userID, page, pageSize int) ([]*ent.Resource, int, error) {
	offset := (page - 1) * pageSize
	resources, err := r.resourceRepo.List(ctx, userID, offset, pageSize)
	if err != nil {
		return nil, 0, err
	}

	count, err := r.resourceRepo.Count(ctx, userID)
	if err != nil {
		return nil, 0, err
	}

	return resources, count, nil
}

func (r *Resource) Get(ctx context.Context, userID, id int) (*ent.Resource, error) {
	rc, err := r.resourceRepo.Get(ctx, userID, id)
	if err != nil {
		return nil, err
	}

	return rc, nil
}

func (r *Resource) Create(
	ctx context.Context,
	userID int,
	name string,
	provider string,
	config map[string]any,
) (*ent.Resource, *ent.Task, error) {
	rc, err := r.resourceRepo.Create(
		ctx,
		userID,
		name,
		provider,
		config,
		"",
	)
	if err != nil {
		return nil, nil, err
	}

	payload := make(map[string]any)
	payload["name"] = name
	payload["provider"] = provider
	payload["config"] = config

	t, err := r.taskRepo.Create(
		ctx,
		userID,
		rc.ID,
		task.TypeCREATE_RESOURCE,
		payload,
	)
	if err != nil {
		return nil, nil, err
	}
	return rc, t, nil
}

func (r *Resource) Delete(ctx context.Context, userID, id int) error {
	return r.resourceRepo.Delete(ctx, userID, id)
}

func (r *Resource) Start(ctx context.Context, userID, id int) error {
	return r.resourceRepo.ResetStatus(ctx, userID, id, resource.StatusSTARTING)
}

func (r *Resource) Stop(ctx context.Context, userID, id int) error {
	return r.resourceRepo.ResetStatus(ctx, userID, id, resource.StatusSTOPPING)
}

func (r *Resource) Restart(ctx context.Context, userID, id int) error {
	return r.resourceRepo.ResetStatus(ctx, userID, id, resource.StatusRESTARTING)
}
