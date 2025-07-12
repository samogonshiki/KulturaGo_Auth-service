package repo_struct

type VisitPoint struct {
	Date  string `db:"date"`
	Count int    `db:"count"`
}

type RatingPoint struct {
	Date  string  `db:"date"`
	Value float64 `db:"value"`
}

type ProfileDB struct {
	UserID          int64  `db:"user_id"`
	FullName        string `db:"full_name"`
	About           string `db:"about"`
	Email           string `db:"email"`
	Avatar          string `db:"avatar"`
	City            string `db:"city"`
	Phone           string `db:"phone"`
	Birthday        string `db:"birthday"`
	TwoFAEnabled    bool   `db:"two_fa_enabled"`
	LoginAlerts     bool   `db:"login_alerts"`
	AllowNewDevices bool   `db:"allow_new_devices"`
}
