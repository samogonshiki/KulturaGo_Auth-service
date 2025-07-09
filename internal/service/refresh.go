package service

import (
	"context"
	"errors"
	"time"
)

func (s *Service) saveRefresh(ctx context.Context, token string) {
	_ = s.rtStore.Save(ctx, token, time.Duration(s.mgr.RefreshTTLSeconds())*time.Second)
}

func (s *Service) Refresh(ctx context.Context, old string) (string, string, error) {
	if ok, _ := s.rtStore.IsActive(ctx, old); !ok {
		return "", "", errors.New("refresh expired")
	}

	cls, err := s.mgr.Parse(old)
	if err != nil {
		return "", "", err
	}

	tks, err := s.mgr.Generate(cls.UserID)
	if err != nil {
		return "", "", err
	}
	s.saveRefresh(ctx, tks.RefreshToken)
	_ = s.rtStore.Revoke(ctx, old)
	return tks.AccessToken, tks.RefreshToken, nil
}

func (s *Service) RevokeAccess(ctx context.Context, jti string) {
	_ = s.rtStore.BlacklistAccess(ctx, jti, time.Hour)
}

func (s *Service) AccessAllowed(ctx context.Context, jti string) bool {
	ok, _ := s.rtStore.IsAccessAllowed(ctx, jti)
	return ok
}
