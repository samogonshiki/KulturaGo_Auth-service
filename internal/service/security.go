package service

import (
	"context"
	"kulturago/auth-service/internal/custom_err"
)

type SecuritySetting struct {
	Key     string `json:"key"`
	Title   string `json:"title"`
	Enabled bool   `json:"enabled"`
}

func (s *Service) Security(ctx context.Context, uid int64) ([]SecuritySetting, error) {
	u, err := s.repo.ByID(ctx, uid)
	if err != nil {
		return nil, err
	}

	return []SecuritySetting{
		{"twoFA", "Двухфакторная аутентификация", u.TwoFAEnabled},
		{"loginAlerts", "Уведомления о входе", u.LoginAlerts},
		{"allowNewDevices", "Новые устройства", u.AllowNewDevices},
	}, nil
}

func (s *Service) ToggleSecurity(ctx context.Context, uid int64, key string, en bool) error {
	return s.repo.UpdateSecurityFlag(ctx, uid, key, en)
}

func (s *Service) ChangePassword(ctx context.Context, uid int64, old, new string) error {
	u, err := s.repo.ByID(ctx, uid)
	if err != nil {
		return err
	}
	if !verify(old, u.PasswordHash) {
		return custom_err.ErrInvalidCreds
	}
	return s.repo.UpdatePassword(ctx, uid, hash(new, salt()))
}
