package repos

import (
	"context"
	"devcloud/ent"
	"devcloud/ent/task"
	"time"
)

type Task struct {
	task *ent.TaskClient
}

func NewTask(db *ent.Client) *Task {
	return &Task{task: db.Task}
}

func (t *Task) Get(ctx context.Context, userID, id int) (*ent.Task, error) {
	return t.task.
		Query().
		Where(
			task.UserID(userID),
			task.ResourceID(id),
		).
		Only(ctx)
}

func (t *Task) Create(
	ctx context.Context,
	userID int,
	resourceID int,
	_type task.Type,
	payload map[string]any,
) (*ent.Task, error) {
	return t.task.
		Create().
		SetUserID(userID).
		SetResourceID(resourceID).
		SetType(_type).
		SetStatus(task.StatusPENDING).
		SetPayload(payload).
		Save(ctx)
}

func (t *Task) Claim(ctx context.Context) (*ent.Task, error) {
	return t.task.
		Query().
		Where(
			task.StatusEQ(task.StatusPENDING),
			task.NextRunAtLTE(time.Now()),
		).
		Order(ent.Asc(task.FieldNextRunAt)).
		Only(ctx)
}
