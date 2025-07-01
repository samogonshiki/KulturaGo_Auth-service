package auth_struct

type SignUpReq struct {
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	Password string `json:"password"`
}
type SignUpResp struct {
	UserID int64 `json:"user_id"`
}

type SignInReq struct{ Email, Password string }

type RefreshReq struct {
	Refresh string `json:"refresh_token"`
}
type RefreshResp struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

type AccessResp struct {
	UserID int64 `json:"user_id"`
	Exp    int64 `json:"exp"`
}
type LogoutReq struct {
	Refresh string `json:"refresh_token"`
}

type ProfileResp struct {
	UserID   int64  `json:"user_id"`
	FullName string `json:"full_name"`
	About    string `json:"about"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar"`
	City     string `json:"city"`
	Phone    string `json:"phone"`
	Birthday string `json:"birthday"`
}

type ProfileReq struct {
	FullName string `json:"full_name"`
	About    string `json:"about"`
	Avatar   string `json:"avatar"`
	City     string `json:"city"`
	Phone    string `json:"phone"`
	Birthday string `json:"birthday"`
}

type AvatarPutResp struct {
	PutURL    string `json:"put_url"`
	PublicURL string `json:"public_url"`
}
