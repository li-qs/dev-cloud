package service

import (
	"context"
	"devcloud/ent"
	"devcloud/repo"
)

type Task struct {
	repo *repo.Repo
}

func NewTask(repo *repo.Repo) *Task {
	return &Task{repo: repo}
}

func (t *Task) GetTask(ctx context.Context, userID, id int) (*ent.Task, error) {
	return t.repo.Task.Get(ctx, userID, id)
}
