package storage

import (
	"database/sql"
	"devcloud/config"
	"devcloud/ent"
	entsql "entgo.io/ent/dialect/sql"

	_ "github.com/lib/pq"
)

// NewPostgres 返回 ent client 以及其底层 *sql.DB（共享同一连接池，用于就绪检查等场景）。
func NewPostgres(cfg *config.Config) (*ent.Client, *sql.DB, error) {
	db, err := sql.Open("postgres", cfg.Postgres)
	if err != nil {
		return nil, nil, err
	}

	drv := entsql.OpenDB("postgres", db)
	client := ent.NewClient(ent.Driver(drv))

	return client, db, nil
}
