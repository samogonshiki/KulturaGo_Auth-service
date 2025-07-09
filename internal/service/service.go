package service

import (
	"context"

	"kulturago/auth-service/internal/domain"
	"kulturago/auth-service/internal/kafka"
	"kulturago/auth-service/internal/redis"
	rp "kulturago/auth-service/internal/repository/repo_struct"
	"kulturago/auth-service/internal/storage"
	"kulturago/auth-service/internal/tokens"
)

type Repository interface {
	ByEmail(ctx context.Context, email string) (*domain.User, error)
	ByProvider(ctx context.Context, prov, pid string) (*domain.User, error)
	Create(ctx context.Context, u *domain.User) error
	CreateBlankProfile(ctx context.Context, uid int64) error
	GetProfileFull(ctx context.Context, uid int64) (rp.ProfileDB, error)
	UpdateProfile(ctx context.Context, p rp.ProfileDB) error
}

type Service struct {
	repo    Repository
	kafka   *kafka.Producer
	mgr     *tokens.Manager
	rtStore *redis.RefreshStore
	store   *storage.S3
}

func New(repo Repository, prod *kafka.Producer, mgr *tokens.Manager,
	rt *redis.RefreshStore, st *storage.S3) *Service {
	return &Service{repo, prod, mgr, rt, st}
}
