package worker

import (
	"context"
	"devcloud/repo"
	"devcloud/repo/repos"
)

type TaskFetcher struct {
	repo *repo.Repo
}

func NewTaskFetcher(repo *repo.Repo) *TaskFetcher {
	return &TaskFetcher{repo: repo}
}

func (t *TaskFetcher) fetch(ctx context.Context) (*repos.ClaimTaskResult, error) {
	return t.repo.Task.Claim(ctx)
}
