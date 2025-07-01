package service

import (
	st "kulturago/auth-service/internal/handler/http/auth_struct"
	repo "kulturago/auth-service/internal/repository/repo_struct"
)

func ToResp(p repo.ProfileDB) st.ProfileResp {
	return st.ProfileResp{
		UserID:   p.UserID,
		FullName: p.FullName,
		About:    p.About,
		Email:    p.Email,
		Avatar:   p.Avatar,
		City:     p.City,
		Phone:    p.Phone,
		Birthday: p.Birthday,
	}
}

func ToDB(uid int64, in st.ProfileReq, email string) repo.ProfileDB {
	return repo.ProfileDB{
		UserID:   uid,
		Email:    email,
		FullName: in.FullName,
		About:    in.About,
		Avatar:   in.Avatar,
		City:     in.City,
		Phone:    in.Phone,
		Birthday: in.Birthday,
	}
}
