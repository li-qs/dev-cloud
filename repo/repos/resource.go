package repos

import (
	"context"
	"devcloud/ent"
	"devcloud/ent/resource"
	"devcloud/ent/task"
)

type Resource struct {
	db       *ent.Client
	resource *ent.ResourceClient
}

func NewResource(db *ent.Client) *Resource {
	return &Resource{db: db, resource: db.Resource}
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
			resource.ID(id),
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
	image string,
	config map[string]any,
	_type task.Type,
	payload map[string]any,
) (*ent.Resource, *ent.Task, error) {
	tx, err := r.db.Tx(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	rsc, err := tx.Resource.
		Create().
		SetUserID(userID).
		SetName(name).
		SetProvider(provider).
		SetImage(image).
		SetStatus(resource.StatusCREATING).
		SetConfig(config).
		Save(ctx)
	if err != nil {
		return nil, nil, err
	}

	t, err := tx.Task.
		Create().
		SetUserID(userID).
		SetResourceID(rsc.ID).
		SetType(_type).
		SetStatus(task.StatusPENDING).
		SetPayload(payload).
		Save(ctx)
	if err != nil {
		return nil, nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, nil, err
	}

	return rsc, t, nil
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
