package redis

import (
	"context"

	rds "github.com/redis/go-redis/v9"
)

type Store struct{ *rds.Client }

func New(addr string) (*Store, error) {
	r := rds.NewClient(&rds.Options{Addr: addr})
	return &Store{r}, nil
}

func (s *Store) Set(ctx context.Context, k, v string) error {
	return s.Client.Set(ctx, k, v, 0).Err()
}
