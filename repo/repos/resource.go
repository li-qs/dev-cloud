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
			resource.StatusNEQ(resource.StatusDELETED),
		).
		Count(ctx)
}

func (r *Resource) Get(ctx context.Context, id int) (*ent.Resource, error) {
	return r.resource.
		Query().
		Where(
			resource.ID(id),
			resource.StatusNEQ(resource.StatusDELETED),
		).
		Only(ctx)
}

func (r *Resource) GetByUser(ctx context.Context, userID, id int) (*ent.Resource, error) {
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

func (r *Resource) DeleteByUser(ctx context.Context, userID, id int) error {
	return r.SetStatusByUser(ctx, userID, id, resource.StatusDELETING)
}

func (r *Resource) SetStatusByUser(ctx context.Context, userID, id int, status resource.Status) error {
	return r.resource.
		UpdateOneID(id).
		Where(
			resource.UserID(userID),
			resource.StatusNEQ(resource.StatusDELETED),
		).
		SetStatus(status).
		Exec(ctx)
}

func (r *Resource) MarkCreated(ctx context.Context, id int, runtimeID string) error {
	return r.resource.
		UpdateOneID(id).
		Where(
			resource.StatusEQ(resource.StatusCREATING),
		).
		SetStatus(resource.StatusRUNNING).
		SetRuntimeID(runtimeID).
		Exec(ctx)
}

func (r *Resource) setStatus(ctx context.Context, id int, from, to resource.Status) error {
	return r.resource.
		UpdateOneID(id).
		Where(
			resource.StatusEQ(from),
		).
		SetStatus(to).
		Exec(ctx)
}

func (r *Resource) MarkFailed(ctx context.Context, id int, from resource.Status) error {
	return r.setStatus(
		ctx,
		id,
		from,
		resource.StatusFAILED,
	)
}

func (r *Resource) MarkRunning(ctx context.Context, id int, from resource.Status) error {
	return r.setStatus(
		ctx,
		id,
		from,
		resource.StatusSTARTING,
	)
}

func (r *Resource) MarkStarting(ctx context.Context, id int) error {
	return r.setStatus(
		ctx,
		id,
		resource.StatusSTOPPED,
		resource.StatusSTARTING,
	)
}

func (r *Resource) MarkRestarting(ctx context.Context, id int) error {
	return r.setStatus(
		ctx,
		id,
		resource.StatusRUNNING,
		resource.StatusRESTARTING,
	)
}

func (r *Resource) MarkStopping(ctx context.Context, id int) error {
	return r.setStatus(
		ctx,
		id,
		resource.StatusRUNNING,
		resource.StatusSTOPPING,
	)
}

func (r *Resource) MarkStopped(ctx context.Context, id int) error {
	return r.setStatus(
		ctx,
		id,
		resource.StatusSTOPPING,
		resource.StatusSTOPPED,
	)
}

func (r *Resource) MarkDeleting(ctx context.Context, id int) error {
	return r.setStatus(
		ctx,
		id,
		resource.StatusSTOPPED,
		resource.StatusDELETING,
	)
}

func (r *Resource) MarkDeleted(ctx context.Context, id int) error {
	return r.setStatus(
		ctx,
		id,
		resource.StatusDELETING,
		resource.StatusDELETED,
	)
}
