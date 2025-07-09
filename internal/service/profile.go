package service

import (
	"context"
	"errors"
	"kulturago/auth-service/internal/repository"
	rp "kulturago/auth-service/internal/repository/repo_struct"
	"log"
)

func (s *Service) Profile(ctx context.Context, uid int64) (rp.ProfileDB, error) {
	pr, err := s.repo.GetProfileFull(ctx, uid)
	if errors.Is(err, repository.ErrNotFound) {
		_ = s.repo.CreateBlankProfile(ctx, uid)
		return s.repo.GetProfileFull(ctx, uid)
	}
	return pr, err
}

func (s *Service) SaveProfile(ctx context.Context, p rp.ProfileDB) error {
	return s.repo.UpdateProfile(ctx, p)
}

func (s *Service) GetAvatarPutURL(ctx context.Context, uid int64) (string, string, error) {
	prof, err := s.repo.GetProfileFull(ctx, uid)
	if err != nil {
		return "", "", err
	}
	put, key, _ := s.store.PresignAvatarPut(ctx, uid, prof.Email)
	log.Printf("PUT → %s\nPUBLIC → %s", put, s.store.PublicURL(key))
	return put, s.store.PublicURL(key), nil
}
