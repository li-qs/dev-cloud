package service

import (
	"context"
	"devcloud/ent"
	"devcloud/ent/resource"
	"devcloud/repo"
)

type Task struct {
	repo *repo.Repo
}

func NewTask(repo *repo.Repo) *Task {
	return &Task{repo: repo}
}

func (t *Task) List(ctx context.Context, userID, page, pageSize int) ([]*ent.Task, int, error) {
	offset := (page - 1) * pageSize
	tasks, err := t.repo.Task.List(ctx, userID, offset, pageSize)
	if err != nil {
		return nil, 0, err
	}

	count, err := t.repo.Task.Count(ctx, userID)
	if err != nil {
		return nil, 0, err
	}

	return tasks, count, nil
}

func (t *Task) Get(ctx context.Context, userID, id int) (*ent.Task, error) {
	return t.repo.Task.Get(ctx, userID, id)
}

func (t *Task) CreateResource(
	ctx context.Context,
	userID int,
	name string,
	provider resource.Provider,
	image string,
	config map[string]any,
) (*ent.Resource, *ent.Task, error) {
	payload := make(map[string]any)
	payload["name"] = name
	payload["provider"] = provider
	payload["image"] = image
	payload["config"] = config

	tk, rsc, err := t.repo.Task.Create(
		ctx,
		userID,
		name,
		provider,
		image,
		config,
		payload,
	)
	if err != nil {
		return nil, nil, err
	}

	return rsc, tk, nil
}

func (t *Task) RemoveResource(ctx context.Context, userID, resourceID int) (*ent.Task, error) {
	return t.repo.Task.RemoveResource(
		ctx,
		userID,
		resourceID,
		map[string]any{},
	)
}

func (t *Task) StartResource(ctx context.Context, userID, resourceID int) (*ent.Task, error) {
	return t.repo.Task.StartResource(
		ctx,
		userID,
		resourceID,
		map[string]any{},
	)
}

func (t *Task) StopResource(ctx context.Context, userID, resourceID int) (*ent.Task, error) {
	return t.repo.Task.StopResource(
		ctx,
		userID,
		resourceID,
		map[string]any{},
	)
}

func (t *Task) RestartResource(ctx context.Context, userID, resourceID int) (*ent.Task, error) {
	return t.repo.Task.RestartResource(
		ctx,
		userID,
		resourceID,
		map[string]any{},
	)
}
