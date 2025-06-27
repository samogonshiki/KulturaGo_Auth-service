package http

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

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

func setCookie(w http.ResponseWriter, name, val string, maxAge int, path string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    val,
		Path:     path,
		MaxAge:   maxAge,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   false,
	})
}

func clearCookie(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{
		Name:   name,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
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
	var in struct{ Email, Password string }
	if json.NewDecoder(r.Body).Decode(&in) != nil {
		http.Error(w, "bad json", 400)
		return
	}

	acc, ref, err := h.svc.SignIn(r.Context(), in.Email, in.Password)
	if err != nil {
		http.Error(w, err.Error(), 401)
		return
	}

	setCookie(w, "access_token", acc, int((15 * time.Minute).Seconds()), "/")
	setCookie(w, "refresh_token", ref, int((30 * 24 * time.Hour).Seconds()), "/api/v1/auth")

	w.WriteHeader(http.StatusNoContent)
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
	c, err := r.Cookie("refresh_token")
	if err != nil || c.Value == "" {
		http.Error(w, "no refresh token", 401)
		return
	}

	acc, ref, err := h.svc.Refresh(r.Context(), c.Value)
	if err != nil {
		http.Error(w, err.Error(), 401)
		return
	}

	setCookie(w, "access_token", acc, int((15 * time.Minute).Seconds()), "/")
	setCookie(w, "refresh_token", ref, int((30 * 24 * time.Hour).Seconds()), "/api/v1/auth")

	w.WriteHeader(http.StatusNoContent)
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
	clearCookie(w, "access_token")
	clearCookie(w, "refresh_token")
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

// GET /api/v1/profile
func (h *AuthHandler) Profile(w http.ResponseWriter, r *http.Request) {
	uid, _ := middleware.FromCtx(r.Context())

	pdb, err := h.svc.Profile(r.Context(), uid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := service.ToResp(pdb) // ← ProfileResp
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// PUT /api/v1/profile
func (h *AuthHandler) SaveProfile(w http.ResponseWriter, r *http.Request) {
	uid, _ := middleware.FromCtx(r.Context())

	var in st.ProfileReq
	if json.NewDecoder(r.Body).Decode(&in) != nil {
		http.Error(w, "bad json", 400)
		return
	}
	cur, _ := h.svc.Profile(r.Context(), uid)

	pdb := service.ToDB(uid, in, cur.Email)
	if err := h.svc.SaveProfile(r.Context(), pdb); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	updated, _ := h.svc.Profile(r.Context(), uid)
	_ = json.NewEncoder(w).Encode(service.ToResp(updated))
}
