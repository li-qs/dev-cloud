package repo

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
			resource.StatusNEQ(resource.StatusDELETED),
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
		).
		Count(ctx)
}

func (r *Resource) Get(ctx context.Context, userID, id int) (*ent.Resource, error) {
	return r.resource.
		Query().
		Where(
			resource.UserID(userID),
			resource.StatusNEQ(resource.StatusDELETED),
		).
		Only(ctx)
}

func (r *Resource) Create(
	ctx context.Context,
	userID int,
	name string,
	provider string,
	config map[string]any,
	credential string,
) (*ent.Resource, error) {
	return r.resource.
		Create().
		SetUserID(userID).
		SetName(name).
		SetProvider(provider).
		SetStatus(resource.StatusCREATING).
		SetConfig(config).
		SetCredential(credential).
		Save(ctx)
}

func (r *Resource) ResetStatus(ctx context.Context, userID, id int, status resource.Status) error {
	return r.resource.
		Update().
		Where(
			resource.ID(id),
			resource.UserID(userID),
			resource.StatusNEQ(resource.StatusDELETED),
		).
		SetStatus(status).
		Exec(ctx)
}

func (r *Resource) Delete(ctx context.Context, userID, id int) error {
	return r.resource.
		Update().
		Where(
			resource.ID(id),
			resource.UserID(userID),
		).
		SetStatus(resource.StatusDELETING).
		Exec(ctx)
}
