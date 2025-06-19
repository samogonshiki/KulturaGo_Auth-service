package http

import (
	"encoding/json"
	"net/http"
	"strings"

	st "kulturago/auth-service/internal/handler/http/auth_struct"
	"kulturago/auth-service/internal/middleware"
	"kulturago/auth-service/internal/service"
	"kulturago/auth-service/internal/tokens"
)

type AuthHandler struct {
	svc *service.Service
	mgr *tokens.Manager
}

func NewAuthHandler(s *service.Service, m *tokens.Manager) *AuthHandler {
	return &AuthHandler{s, m}
}

// @Summary      Регистрация
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        payload body      signUpReq  true  "nickname, email, password"
// @Success      200     {object}  signUpResp
// @Failure      409     {string}  string     "user exists"
// @Router       /api/v1/auth/signup [post]
func (h *AuthHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	var in st.SignUpReq
	if json.NewDecoder(r.Body).Decode(&in) != nil ||
		in.Nickname == "" || in.Email == "" || len(in.Password) < 6 {
		http.Error(w, "validation failed", 422)
		return
	}
	u, err := h.svc.SignUp(r.Context(), in.Email, in.Nickname, in.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	_ = json.NewEncoder(w).Encode(st.SignUpResp{UserID: u.ID})
}

// @Summary      Логин
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        payload body      signInReq true "email, password"
// @Success      200     {object}  map[string]string "access_token / refresh_token"
// @Failure      401     {string}  string            "invalid credentials"
// @Router       /api/v1/auth/signin [post]
func (h *AuthHandler) SignIn(w http.ResponseWriter, r *http.Request) {
	var in st.SignInReq
	if json.NewDecoder(r.Body).Decode(&in) != nil {
		http.Error(w, "bad json", 400)
		return
	}
	access, refresh, err := h.svc.SignIn(r.Context(), in.Email, in.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]string{
		"access_token":  access,
		"refresh_token": refresh,
	})
}

// @Summary      Обновление токенов
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        payload body      refreshReq true "refresh_token"
// @Success      200     {object}  refreshResp
// @Failure      401     {string}  string "invalid / revoked token"
// @Router       /api/v1/auth/refresh [post]
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var in st.RefreshReq
	if json.NewDecoder(r.Body).Decode(&in) != nil || in.Refresh == "" {
		http.Error(w, "missing refresh_token", 400)
		return
	}
	cls, err := h.mgr.Parse(in.Refresh)
	if err != nil || cls.RegisteredClaims.Subject != "refresh" {
		http.Error(w, "invalid token", 401)
		return
	}
	if !h.svc.RefreshActive(r.Context(), in.Refresh) {
		http.Error(w, "token revoked", 401)
		return
	}
	tks, _ := h.mgr.Generate(cls.UserID)
	h.svc.SaveRefresh(r.Context(), tks.RefreshToken, tks.RefreshExpiresIn)
	_ = json.NewEncoder(w).Encode(st.RefreshResp{
		AccessToken:  tks.AccessToken,
		RefreshToken: tks.RefreshToken,
		ExpiresIn:    tks.ExpiresIn,
	})
}

// @Summary      Проверка access-токена
// @Tags         auth
// @Security     Bearer
// @Produce      json
// @Success      200 {object} accessResp
// @Failure      401 {string} string "invalid / revoked token"
// @Router       /api/v1/auth/access [get]
func (h *AuthHandler) Access(w http.ResponseWriter, r *http.Request) {
	bearer := r.Header.Get("Authorization")
	if !strings.HasPrefix(bearer, "Bearer ") {
		http.Error(w, "no bearer", 401)
		return
	}
	token := strings.TrimPrefix(bearer, "Bearer ")
	cls, err := h.mgr.Parse(token)
	if err != nil || !h.svc.AccessAllowed(r.Context(), cls.ID) {
		http.Error(w, "invalid token", 401)
		return
	}
	_ = json.NewEncoder(w).Encode(st.AccessResp{
		UserID: cls.UserID,
		Exp:    cls.ExpiresAt.Unix(),
	})
}

// @Summary      Logout (отзыв refresh-токена)
// @Tags         auth
// @Accept       json
// @Param        payload body logoutReq true "refresh_token"
// @Success      204  "no content"
// @Router       /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var in st.LogoutReq
	if json.NewDecoder(r.Body).Decode(&in) != nil || in.Refresh == "" {
		http.Error(w, "missing refresh_token", 400)
		return
	}
	if err := h.svc.RevokeRefresh(r.Context(), in.Refresh); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// @Summary      Профиль текущего пользователя
// @Tags         auth
// @Security     Bearer
// @Produce      json
// @Success      200 {object} map[string]int64
// @Router       /api/v1/me [get]
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	uid, _ := middleware.FromCtx(r.Context())
	_ = json.NewEncoder(w).Encode(map[string]int64{"user_id": uid})
}
