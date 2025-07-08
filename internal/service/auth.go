package service

import (
	"context"
	Cerr "kulturago/auth-service/internal/custom_err"
	"kulturago/auth-service/internal/domain"
	"kulturago/auth-service/internal/repository"

	"github.com/google/uuid"
)

func (s *Service) SignUp(ctx context.Context, email, nick, pwd string) (*domain.User, error) {
	if _, err := s.repo.ByEmail(ctx, email); err == nil {
		return nil, Cerr.ErrExists
	}
	u := &domain.User{Email: email, Nickname: nick, PasswordHash: hash(pwd, salt())}
	if err := s.repo.Create(ctx, u); err != nil {
		return nil, err
	}
	_ = s.repo.CreateBlankProfile(ctx, u.ID)
	s.kafka.SendRaw("user.created", "", []byte(`{"id":"`+uuid.NewString()+`"}`))
	return u, nil
}

func (s *Service) SignIn(ctx context.Context, email, pwd string) (string, string, error) {
	u, err := s.repo.ByEmail(ctx, email)
	if err != nil || !verify(pwd, u.PasswordHash) {
		return "", "", Cerr.ErrInvalidCreds
	}
	tks, err := s.mgr.Generate(u.ID)
	if err != nil {
		return "", "", err
	}
	s.saveRefresh(ctx, tks.RefreshToken)
	return tks.AccessToken, tks.RefreshToken, nil
}

func (s *Service) SocialLogin(ctx context.Context, prov, pid, email string) (string, string, error) {
	u, err := s.repo.ByProvider(ctx, prov, pid)
	if err == repository.ErrNotFound {
		u = &domain.User{Email: email, Provider: prov, ProviderID: pid}
		if err = s.repo.Create(ctx, u); err != nil {
			return "", "", err
		}
	}
	tks, err := s.mgr.Generate(u.ID)
	if err != nil {
		return "", "", err
	}
	s.saveRefresh(ctx, tks.RefreshToken)
	return tks.AccessToken, tks.RefreshToken, nil
}
