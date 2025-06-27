package repo_struct

type ProfileDB struct {
	UserID   int64  `db:"user_id"`
	FullName string `db:"full_name"`
	About    string `db:"about"`
	Email    string `db:"email"`
	Avatar   string `db:"avatar"`
	City     string `db:"city"`
	Phone    string `db:"phone"`
	Birthday string `db:"birthday"`
}
