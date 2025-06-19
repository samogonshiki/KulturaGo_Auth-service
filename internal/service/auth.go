package service

import (
	"context"
	"crypto/rand"
	"errors"
	"kulturago/auth-service/internal/domain"
	"kulturago/auth-service/internal/kafka"
	"kulturago/auth-service/internal/redis"
	"kulturago/auth-service/internal/repository"
	"kulturago/auth-service/internal/tokens"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/argon2"
)

type Service struct {
	repo    *repository.PG
	kafka   *kafka.Producer
	mgr     *tokens.Manager
	rtStore *redis.RefreshStore
}

var (
	ErrExists       = errors.New("user exists")
	ErrInvalidCreds = errors.New("invalid credentials")
)

func New(repo *repository.PG, prod *kafka.Producer, mgr *tokens.Manager,
	rt *redis.RefreshStore) *Service {
	return &Service{repo, prod, mgr, rt}
}

func salt() []byte { b := make([]byte, 16); rand.Read(b); return b }
func hash(pwd string, s []byte) []byte {
	return append(s, argon2.IDKey([]byte(pwd), s, 1, 64*1024, 4, 32)...)
}
func verify(pwd string, h []byte) bool {
	s := h[:16]
	return string(h[16:]) == string(argon2.IDKey([]byte(pwd), s, 1, 64*1024, 4, 32))
}

func (s *Service) SignUp(ctx context.Context, email, pwd, password string) (*domain.User, error) {
	if _, err := s.repo.ByEmail(ctx, email); err == nil {
		return nil, ErrExists
	}
	u := &domain.User{
		Email:        email,
		PasswordHash: hash(pwd, salt()),
	}
	if err := s.repo.Create(ctx, u); err != nil {
		return nil, err
	}
	s.kafka.SendRaw("user.created", "", []byte(`{"id":"`+uuid.NewString()+`"}`))
	return u, nil
}

func (s *Service) SignIn(ctx context.Context, email, pwd string) (string, string, error) {
	u, err := s.repo.ByEmail(ctx, email)
	if err != nil || !verify(pwd, u.PasswordHash) {
		return "", "", ErrInvalidCreds
	}
	tks, _ := s.mgr.Generate(u.ID)
	return tks.AccessToken, tks.RefreshToken, nil
}

func (s *Service) SocialLogin(ctx context.Context, provider, pid, email string) (string, string, error) {
	u, err := s.repo.ByProvider(ctx, provider, pid)
	if err == repository.ErrNotFound {
		u = &domain.User{
			Email:      email,
			Provider:   provider,
			ProviderID: pid,
		}
		if err = s.repo.Create(ctx, u); err != nil {
			return "", "", err
		}
	}
	tk, _ := s.mgr.Generate(u.ID)
	return tk.AccessToken, tk.RefreshToken, nil
}

func (s *Service) SaveRefresh(ctx context.Context, token string, secs int64) {
	_ = s.rtStore.Save(ctx, token, time.Duration(secs)*time.Second)
}

func (s *Service) RefreshActive(ctx context.Context, token string) bool {
	ok, _ := s.rtStore.IsActive(ctx, token)
	return ok
}

func (s *Service) RevokeRefresh(ctx context.Context, token string) error {
	return s.rtStore.Revoke(ctx, token)
}

func (s *Service) RevokeAccess(ctx context.Context, jti string, secs int64) {
	_ = s.rtStore.BlacklistAccess(ctx, jti, time.Duration(secs)*time.Second)
}

func (s *Service) AccessAllowed(ctx context.Context, jti string) bool {
	ok, _ := s.rtStore.IsAccessAllowed(ctx, jti)
	return ok
}
