package service

import (
	st "kulturago/auth-service/internal/handler/http/auth_struct"
	rp "kulturago/auth-service/internal/repository/repo_struct"
)

func ToResp(p rp.ProfileDB) st.ProfileResp {
	return st.ProfileResp{
		UserID:          p.UserID,
		FullName:        p.FullName,
		About:           p.About,
		Email:           p.Email,
		Avatar:          p.Avatar,
		City:            p.City,
		Phone:           p.Phone,
		Birthday:        p.Birthday,
		TwoFAEnabled:    p.TwoFAEnabled,
		LoginAlerts:     p.LoginAlerts,
		AllowNewDevices: p.AllowNewDevices,
	}
}

func ToDB(uid int64, in st.ProfileReq, email string) rp.ProfileDB {
	return rp.ProfileDB{
		UserID:          uid,
		Email:           email,
		FullName:        in.FullName,
		About:           in.About,
		Avatar:          in.Avatar,
		City:            in.City,
		Phone:           in.Phone,
		Birthday:        in.Birthday,
		TwoFAEnabled:    in.TwoFAEnabled,
		LoginAlerts:     in.LoginAlerts,
		AllowNewDevices: in.AllowNewDevices,
	}
}
