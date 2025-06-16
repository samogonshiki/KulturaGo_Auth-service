package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"kulturago/auth-service/internal/domain"
)

type PG struct{ db *pgxpool.Pool }

func New(dsn string) (*PG, error) {
	pool, err := pgxpool.New(context.Background(), dsn)
	return &PG{pool}, err
}

var ErrNotFound = errors.New("not found")

func (p *PG) Tx(ctx context.Context, fn func(pgx.Tx) error) error {
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	return tx.Commit(ctx)
}

func (p *PG) Create(ctx context.Context, u *domain.User) error {
	return p.db.QueryRow(ctx,
		`INSERT INTO users(email,password_hash,provider,provider_id)
		 VALUES($1,$2,$3,$4) RETURNING id`,
		u.Email, u.PasswordHash, u.Provider, u.ProviderID,
	).Scan(&u.ID)
}

func (p *PG) ByEmail(ctx context.Context, email string) (*domain.User, error) {
	var u domain.User
	err := p.db.QueryRow(ctx,
		`SELECT id,email,password_hash,provider,provider_id,created_at
		   FROM users WHERE email=$1`, email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash,
		&u.Provider, &u.ProviderID, &u.CreatedAt)
	if err != nil {
		return nil, ErrNotFound
	}
	return &u, nil
}

func (p *PG) ByProvider(ctx context.Context, provider, pid string) (*domain.User, error) {
	var u domain.User
	err := p.db.QueryRow(ctx,
		`SELECT id,email,password_hash,provider,provider_id,created_at
		   FROM users WHERE provider=$1 AND provider_id=$2`,
		provider, pid,
	).Scan(&u.ID, &u.Email, &u.PasswordHash,
		&u.Provider, &u.ProviderID, &u.CreatedAt)
	if err != nil {
		return nil, ErrNotFound
	}
	return &u, nil
}
