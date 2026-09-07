package repos

import (
	"context"
	"devcloud/ent"
	"devcloud/ent/task"
	"fmt"
	"time"
)

type Task struct {
	db   *ent.Client
	task *ent.TaskClient
}

func NewTask(db *ent.Client) *Task {
	return &Task{db: db, task: db.Task}
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
			" WHERE `status`=$1 AND next_run_at<=$2"+
			" ORDER BY next_run_at ASC"+
			" LIMIT 1"+
			" FOR UPDATE SKIP LOCKED",
		task.StatusPENDING,
		time.Now(),
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
		"UPDATE task SET `status`=$1 WHERE id=$2",
		task.StatusRUNNING,
		task.ID,
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
	return t.task.
		UpdateOneID(id).
		SetStatus(task.StatusSUCCESS).
		Exec(ctx)
}

func (t *Task) SetFailed(ctx context.Context, id int, errMsg string) error {
	return t.task.
		UpdateOneID(id).
		SetStatus(task.StatusFAILED).
		SetErrorMessage(errMsg).
		Exec(ctx)
}
