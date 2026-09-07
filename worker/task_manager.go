package worker

import (
	"context"
	"devcloud/repo"
	"devcloud/repo/repos"
)

type TaskManager struct {
	repo *repo.Repo
}

func newTaskManager(repo *repo.Repo) *TaskManager {
	return &TaskManager{repo: repo}
}

func (t *TaskManager) fetch(ctx context.Context) (*repos.ClaimTaskResult, error) {
	return t.repo.Task.Claim(ctx)
}

func (t *TaskManager) success(ctx context.Context, id int) error {
	return t.repo.Task.SetSuccess(ctx, id)
}

func (t *TaskManager) fail(ctx context.Context, id int, errMsg string) error {
	return t.repo.Task.SetFailed(ctx, id, errMsg)
}
