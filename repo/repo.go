package repo

import (
	"devcloud/ent"
	"devcloud/repo/repos"
)

type Repo struct {
	User         *repos.User
	RefreshToken *repos.RefreshToken
	Resource     *repos.Resource
	Task         *repos.Task
}

func New(db *ent.Client) *Repo {
	return &Repo{
		User:         repos.NewUser(db),
		RefreshToken: repos.NewToken(db),
		Resource:     repos.NewResource(db),
		Task:         repos.NewTask(db),
	}
}
