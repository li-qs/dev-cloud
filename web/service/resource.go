package service

import (
	"context"
	"devcloud/ent"
	"devcloud/repo"
)

type Resource struct {
	repo *repo.Repo
}

func NewResource(repo *repo.Repo) *Resource {
	return &Resource{repo: repo}
}

func (r *Resource) List(ctx context.Context, userID, page, pageSize int) ([]*ent.Resource, int, error) {
	offset := (page - 1) * pageSize
	resources, err := r.repo.Resource.List(ctx, userID, offset, pageSize)
	if err != nil {
		return nil, 0, err
	}

	count, err := r.repo.Resource.Count(ctx, userID)
	if err != nil {
		return nil, 0, err
	}

	return resources, count, nil
}

func (r *Resource) Get(ctx context.Context, userID, id int) (*ent.Resource, error) {
	rc, err := r.repo.Resource.GetByUser(ctx, userID, id)
	if err != nil {
		return nil, err
	}

	return rc, nil
}
