package worker

import (
	"context"
	"devcloud/ent"
	"devcloud/repo"
)

type TaskFetcher struct {
	repo *repo.Repo
}

func NewTaskFetcker(repo *repo.Repo) *TaskFetcher {
	return &TaskFetcher{repo: repo}
}

func (t *TaskFetcher) fetch(ctx context.Context) (*ent.Task, error) {
	return t.repo.Task.Claim(ctx)
}
