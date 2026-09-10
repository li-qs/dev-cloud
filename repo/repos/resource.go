package repos

import (
	"context"
	"devcloud/ent"
	"devcloud/ent/resource"
)

type Resource struct {
	db *ent.Client
}

func NewResource(db *ent.Client) *Resource {
	return &Resource{db: db}
}

// List 分页返回用户的资源（不含已删除）。
func (r *Resource) List(ctx context.Context, userID, offset, limit int) ([]*ent.Resource, error) {
	return r.db.Resource.
		Query().
		Where(
			resource.UserID(userID),
			resource.StatusNEQ(resource.StatusREMOVED),
		).
		Offset(offset).
		Limit(limit).
		All(ctx)
}

// Count 统计用户的资源总数（不含已删除）。
func (r *Resource) Count(ctx context.Context, userID int) (int, error) {
	return r.db.Resource.
		Query().
		Where(
			resource.UserID(userID),
			resource.StatusNEQ(resource.StatusREMOVED),
		).
		Count(ctx)
}

// Get 按 id 查询资源（worker 使用，不带用户过滤；已删除不可见）。
func (r *Resource) Get(ctx context.Context, id int) (*ent.Resource, error) {
	return r.db.Resource.
		Query().
		Where(
			resource.ID(id),
			resource.StatusNEQ(resource.StatusREMOVED),
		).
		Only(ctx)
}

// GetByUser 按 id + 用户查询资源，用于对外接口的归属校验；非本人资源返回 NotFound。
func (r *Resource) GetByUser(ctx context.Context, userID, id int) (*ent.Resource, error) {
	return r.db.Resource.
		Query().
		Where(
			resource.ID(id),
			resource.UserID(userID),
			resource.StatusNEQ(resource.StatusREMOVED),
		).
		Only(ctx)
}

// SetCreateSuccess 仅当资源仍为 CREATING 时写入 runtime_id 并置为 STOPPED。
// 目标 STOPPED 是因为 Docker Create 只创建容器、不启动；条件更新防止覆盖已变化的资源。
func (r *Resource) SetCreateSuccess(ctx context.Context, id int, runtimeID string) error {
	return r.db.Resource.
		UpdateOneID(id).
		Where(
			resource.StatusEQ(resource.StatusCREATING),
		).
		SetRuntimeID(runtimeID).
		SetStatus(resource.StatusSTOPPED).
		Exec(ctx)
}

// SetCreateFailed 仅当资源仍为 CREATING 时置为 FAILED。
func (r *Resource) SetCreateFailed(ctx context.Context, id int) error {
	return r.db.Resource.
		UpdateOneID(id).
		Where(
			resource.StatusEQ(resource.StatusCREATING),
		).
		SetStatus(resource.StatusFAILED).
		Exec(ctx)
}

// SetStatus 无条件更新资源状态，用于 worker 执行成功/失败后的状态推进。
func (r *Resource) SetStatus(ctx context.Context, id int, status resource.Status) error {
	return r.db.Resource.
		UpdateOneID(id).
		SetStatus(status).
		Exec(ctx)
}
