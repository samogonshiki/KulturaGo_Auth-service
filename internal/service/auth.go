package service

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"errors"
	"time"

	"kulturago/auth-service/internal/domain"
	"kulturago/auth-service/internal/kafka"
	"kulturago/auth-service/internal/redis"
	"kulturago/auth-service/internal/repository"
	rp "kulturago/auth-service/internal/repository/repo_struct"
	"kulturago/auth-service/internal/tokens"

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

func salt() []byte { b := make([]byte, 16); _, _ = rand.Read(b); return b }

func hash(pwd string, s []byte) []byte {
	return append(s, argon2.IDKey([]byte(pwd), s, 1, 64*1024, 4, 32)...)
}
func verify(pwd string, h []byte) bool {
	s := h[:16]
	cmp := argon2.IDKey([]byte(pwd), s, 1, 64*1024, 4, 32)
	return subtle.ConstantTimeCompare(h[16:], cmp) == 1
}

func (s *Service) SignUp(ctx context.Context, email, nickname, pwd string) (*domain.User, error) {
	if _, err := s.repo.ByEmail(ctx, email); err == nil {
		return nil, ErrExists
	}
	u := &domain.User{
		Email:        email,
		Nickname:     nickname,
		PasswordHash: hash(pwd, salt()),
	}
	if err := s.repo.Create(ctx, u); err != nil {
		return nil, err
	}
	s.kafka.SendRaw("user.created", "", []byte(`{"id":"`+uuid.NewString()+`"}`))
	_ = s.repo.CreateBlankProfile(ctx, u.ID)
	return u, nil
}

func (s *Service) SignIn(ctx context.Context, email, pwd string) (string, string, error) {
	u, err := s.repo.ByEmail(ctx, email)
	if err != nil || !verify(pwd, u.PasswordHash) {
		return "", "", ErrInvalidCreds
	}
	tks, err := s.mgr.Generate(u.ID)
	if err != nil {
		return "", "", err
	}

	s.SaveRefresh(ctx, tks.RefreshToken, s.mgr.RefreshTTLSeconds)

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
	tks, err := s.mgr.Generate(u.ID)
	if err != nil {
		return "", "", err
	}

	s.SaveRefresh(ctx, tks.RefreshToken, s.mgr.RefreshTTLSeconds)

	return tks.AccessToken, tks.RefreshToken, nil
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

func (s *Service) Profile(ctx context.Context, uid int64) (rp.ProfileDB, error) {
	return s.repo.GetProfileFull(ctx, uid)
}

func (s *Service) SaveProfile(ctx context.Context, p rp.ProfileDB) error {
	return s.repo.UpdateProfile(ctx, p)
}

func (s *Service) Refresh(ctx context.Context, oldRefresh string) (string, string, error) {
	if !s.RefreshActive(ctx, oldRefresh) {
		return "", "", errors.New("refresh expired")
	}
	cls, err := s.mgr.Parse(oldRefresh)
	if err != nil {
		return "", "", err
	}

	tks, err := s.mgr.Generate(cls.UserID)
	if err != nil {
		return "", "", err
	}

	s.SaveRefresh(ctx, tks.RefreshToken, s.mgr.RefreshTTLSeconds)
	_ = s.RevokeRefresh(ctx, oldRefresh)

	return tks.AccessToken, tks.RefreshToken, nil
}
