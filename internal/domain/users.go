package domain

import "time"

type User struct {
	ID           int64
	Email        string
	PasswordHash []byte
	Provider     string
	ProviderID   string
	CreatedAt    time.Time
}
