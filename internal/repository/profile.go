package repository

import (
	"context"
	"errors"
	"fmt"
	repo "kulturago/auth-service/internal/repository/repo_struct"

	"github.com/jackc/pgx/v5"
)

func (p *PG) CreateBlankProfile(ctx context.Context, uid int64) error {
	_, err := p.db.Exec(ctx, `
		INSERT INTO profiles (user_id)
		VALUES ($1)
		ON CONFLICT (user_id) DO NOTHING`,
		uid)
	return err
}

func (p *PG) GetProfileFull(ctx context.Context, uid int64) (repo.ProfileDB, error) {
	var pr repo.ProfileDB
	const q = `
SELECT u.email,
       COALESCE(p.full_name, u.nickname, '') AS full_name,
       COALESCE(p.about, '')                 AS about,
       COALESCE(p.avatar,'')                 AS avatar,
       COALESCE(p.city,'')                   AS city,
       COALESCE(p.phone,'')                  AS phone,
       COALESCE(to_char(p.birthday,'YYYY-MM-DD'),'') AS birthday,
       u.two_fa_enabled,
       u.login_alerts,
       u.allow_new_devices
  FROM users u
  LEFT JOIN profiles p ON p.user_id = u.id
 WHERE u.id = $1;
`
	err := p.db.QueryRow(ctx, q, uid).Scan(
		&pr.Email, &pr.FullName, &pr.About, &pr.Avatar,
		&pr.City, &pr.Phone, &pr.Birthday,
		&pr.TwoFAEnabled, &pr.LoginAlerts, &pr.AllowNewDevices,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return repo.ProfileDB{}, ErrNotFound
	}
	pr.UserID = uid
	return pr, err
}

func (p *PG) UpdateProfile(ctx context.Context, pr repo.ProfileDB) error {
	_, err := p.db.Exec(ctx, `
INSERT INTO profiles (user_id, full_name, about, avatar, city, phone, birthday)
VALUES ($1,$2,$3,$4,$5,$6,$7)
ON CONFLICT (user_id) DO UPDATE
  SET full_name = EXCLUDED.full_name,
      about     = EXCLUDED.about,
      avatar    = EXCLUDED.avatar,
      city      = EXCLUDED.city,
      phone     = EXCLUDED.phone,
      birthday  = EXCLUDED.birthday;`,
		pr.UserID,
		pr.FullName,
		pr.About,
		pr.Avatar,
		pr.City,
		pr.Phone,
		pr.Birthday,
	)
	return err
}

func (p *PG) UpdateSecurityFlag(ctx context.Context, uid int64, key string, en bool) error {
	col, ok := map[string]string{
		"twoFA":           "two_fa_enabled",
		"loginAlerts":     "login_alerts",
		"allowNewDevices": "allow_new_devices",
	}[key]
	if !ok {
		return fmt.Errorf("unknown security flag %q", key)
	}
	_, err := p.db.Exec(ctx,
		fmt.Sprintf(`UPDATE users SET %s=$2 WHERE id=$1`, col), uid, en)
	return err
}
