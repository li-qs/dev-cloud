package store

import "github.com/redis/go-redis/v9"

type Store struct {
}

func New(rdb *redis.Client) *Store {
	return &Store{}
}
