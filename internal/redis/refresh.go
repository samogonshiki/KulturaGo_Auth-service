package redis

import (
	"context"
	rds "github.com/redis/go-redis/v9"
	"time"
)

type RefreshStore struct {
	r *rds.Client
}

func NewRefresh(r *rds.Client) *RefreshStore { return &RefreshStore{r} }

func (s *RefreshStore) Save(ctx context.Context, token string, ttl time.Duration) error {
	return s.r.Set(ctx, "rt:"+token, 1, ttl).Err()
}

func (s *RefreshStore) IsActive(ctx context.Context, token string) (bool, bool) {
	ok, _ := s.r.Exists(ctx, "rt:"+token).Result()
	return ok == 1, false
}

func (s *RefreshStore) Revoke(ctx context.Context, token string) error {
	return s.r.Del(ctx, "rt:"+token).Err()
}

func (s *RefreshStore) BlacklistAccess(ctx context.Context, jti string, ttl time.Duration) error {
	return s.r.Set(ctx, "blk:"+jti, 1, ttl).Err()
}
func (s *RefreshStore) IsAccessAllowed(ctx context.Context, jti string) (bool, bool) {
	ok, _ := s.r.Exists(ctx, "blk:"+jti).Result()
	return ok == 0, false
}
