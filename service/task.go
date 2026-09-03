package service

import (
	"context"
	"devcloud/ent"
	"devcloud/repo"
)

type Task struct {
	taskRepo *repo.Task
}

func NewTask(taskRepo *repo.Task) *Task {
	return &Task{taskRepo: taskRepo}
}

func (t *Task) GetTask(ctx context.Context, userID, id int) (*ent.Task, error) {
	return t.taskRepo.Get(ctx, userID, id)
}
