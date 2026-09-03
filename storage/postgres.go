package storage

import (
	"devcloud/config"
	"devcloud/ent"
)

func NewPostgres(cfg *config.Config) (*ent.Client, error) {
	return ent.Open("postgres", cfg.Postgres)
}