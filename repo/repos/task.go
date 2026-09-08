package repos

import (
	"context"
	"devcloud/ent"
	"devcloud/ent/task"
	"fmt"
)

type Task struct {
	db *ent.Client
}

func NewTask(db *ent.Client) *Task {
	return &Task{db: db}
}

func (t *Task) Get(ctx context.Context, userID, id int) (*ent.Task, error) {
	return t.db.Task.
		Query().
		Where(
			task.ID(id),
			task.UserID(userID),
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
	return t.db.Task.
		Create().
		SetUserID(userID).
		SetResourceID(resourceID).
		SetType(_type).
		SetStatus(task.StatusPENDING).
		SetPayload(payload).
		Save(ctx)
}

// HasActive 判断该资源是否已有在途任务（PENDING/RUNNING）。
func (t *Task) HasActive(ctx context.Context, resourceID int) (bool, error) {
	n, err := t.db.Task.
		Query().
		Where(
			task.ResourceID(resourceID),
			task.StatusIn(task.StatusPENDING, task.StatusRUNNING),
		).
		Count(ctx)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

type ClaimTaskResult struct {
	ID         int
	ResourceID int
	Type       task.Type
}

func (t *Task) Claim(ctx context.Context) (*ClaimTaskResult, error) {
	tx, err := t.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	rows, err := tx.QueryContext(
		ctx,
		"SELECT"+
			" id,"+
			" resource_id,"+
			" type"+
			" FROM task"+
			" WHERE status=$1"+
			" ORDER BY id"+
			" LIMIT 1"+
			" FOR UPDATE SKIP LOCKED",
		task.StatusPENDING,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var (
		id         int
		resourceID int
		_type      task.Type
	)
	if err := rows.Scan(
		&id,
		&resourceID,
		&_type,
	); err != nil {
		return nil, err
	}

	res, err := tx.ExecContext(
		ctx,
		"UPDATE task SET status=$1 WHERE id=$2",
		task.StatusRUNNING,
		id,
	)
	if err != nil {
		return nil, err
	}

	affectRows, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if affectRows != 1 {
		return nil, fmt.Errorf("update task status affectRows: %d", affectRows)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &ClaimTaskResult{
		ID:         id,
		ResourceID: resourceID,
		Type:       _type,
	}, nil
}

func (t *Task) SetSuccess(ctx context.Context, id int) error {
	return t.db.Task.
		UpdateOneID(id).
		SetStatus(task.StatusSUCCESS).
		Exec(ctx)
}

func (t *Task) SetFailed(ctx context.Context, id int, errMsg string) error {
	return t.db.Task.
		UpdateOneID(id).
		SetStatus(task.StatusFAILED).
		SetErrorMessage(errMsg).
		Exec(ctx)
}
