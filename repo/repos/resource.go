package repos

import (
	"context"
	"devcloud/ent"
	"devcloud/ent/resource"
)

type Resource struct {
	resource *ent.ResourceClient
}

func NewResource(db *ent.Client) *Resource {
	return &Resource{resource: db.Resource}
}

func (r *Resource) List(ctx context.Context, userID, offset, limit int) ([]*ent.Resource, error) {
	return r.resource.
		Query().
		Where(
			resource.UserID(userID),
			resource.StatusNEQ(resource.StatusREMOVED),
		).
		Offset(offset).
		Limit(limit).
		All(ctx)
}

func (r *Resource) Count(ctx context.Context, userID int) (int, error) {
	return r.resource.
		Query().
		Where(
			resource.UserID(userID),
			resource.StatusNEQ(resource.StatusREMOVED),
		).
		Count(ctx)
}

func (r *Resource) Get(ctx context.Context, id int) (*ent.Resource, error) {
	return r.resource.
		Query().
		Where(
			resource.ID(id),
			resource.StatusNEQ(resource.StatusREMOVED),
		).
		Only(ctx)
}

func (r *Resource) GetByUser(ctx context.Context, userID, id int) (*ent.Resource, error) {
	return r.resource.
		Query().
		Where(
			resource.UserID(userID),
			resource.StatusNEQ(resource.StatusREMOVED),
		).
		Only(ctx)
}

func (r *Resource) Create(
	ctx context.Context,
	userID int,
	name string,
	provider resource.Provider,
	config map[string]any,
) (*ent.Resource, error) {
	return r.resource.
		Create().
		SetUserID(userID).
		SetName(name).
		SetProvider(provider).
		SetStatus(resource.StatusCREATING).
		SetConfig(config).
		Save(ctx)
}

func (r *Resource) SetStatusRUNNING(ctx context.Context, id int, runtimeID string) error {
	return r.resource.
		UpdateOneID(id).
		SetRuntimeID(runtimeID).
		SetStatus(resource.StatusRUNNING).
		Exec(ctx)
}

func (r *Resource) SetStatus(ctx context.Context, id int, to resource.Status) error {
	return r.resource.
		UpdateOneID(id).
		SetStatus(to).
		Exec(ctx)
}
