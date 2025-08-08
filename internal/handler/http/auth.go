package http

import (
	"encoding/json"
	"errors"
	"kulturago/auth-service/internal/custom_err"
	st "kulturago/auth-service/internal/handler/http/auth_struct"
	"kulturago/auth-service/internal/middleware"
	"kulturago/auth-service/internal/service"
	"kulturago/auth-service/internal/tokens"
	utl "kulturago/auth-service/internal/util"
	"net/http"
	"strings"
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
	var in struct{ Email, Password string }
	if json.NewDecoder(r.Body).Decode(&in) != nil {
		http.Error(w, "bad json", 400)
		return
	}

	acc, ref, err := h.svc.SignIn(r.Context(), in.Email, in.Password)
	if err != nil {
		if errors.Is(err, custom_err.ErrNotAdmin) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	accessTTL := int(h.mgr.AccessTTLSeconds())
	refreshTTL := int(h.mgr.RefreshTTLSeconds())

	utl.Set(w, "access_token", acc, accessTTL, "/")
	utl.Set(w, "refresh_token", ref, refreshTTL, "/")

	utl.ClearPath(w, "refresh_token", "/api/v1/auth")

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

	accessTTL := int(h.mgr.AccessTTLSeconds())
	refreshTTL := int(h.mgr.RefreshTTLSeconds())

	utl.Set(w, "access_token", acc, accessTTL, "/")
	utl.Set(w, "refresh_token", ref, refreshTTL, "/")

	utl.ClearPath(w, "refresh_token", "/api/v1/auth")

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
	utl.Clear(w, "access_token")
	utl.Clear(w, "refresh_token")
	utl.ClearPath(w, "refresh_token", "/api/v1/auth")
	w.WriteHeader(http.StatusNoContent)
}

// @Summary      Профиль текущего пользователя
// @Tags         auth
// @Security     Bearer
// @Produce      json
// @Success      200 {object} map[string]int64
// @Router       /api/v1/me [get]
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	cls, ok := middleware.FromCtx(r.Context())
	if !ok {
		http.Error(w, "no claims", http.StatusUnauthorized)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]int64{
		"user_id": cls.UserID,
	})
}

func (h *AuthHandler) Profile(w http.ResponseWriter, r *http.Request) {
	cls, ok := middleware.FromCtx(r.Context())
	if !ok {
		http.Error(w, "no claims", http.StatusUnauthorized)
		return
	}

	pdb, err := h.svc.Profile(r.Context(), cls.UserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(service.ToResp(pdb))
}

func (h *AuthHandler) SaveProfile(w http.ResponseWriter, r *http.Request) {
	cls, ok := middleware.FromCtx(r.Context())
	if !ok {
		http.Error(w, "no claims", http.StatusUnauthorized)
		return
	}

	var in st.ProfileReq
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	cur, err := h.svc.Profile(r.Context(), cls.UserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if in.FullName != "" {
		cur.FullName = in.FullName
	}
	if in.About != "" {
		cur.About = in.About
	}
	if in.Avatar != "" {
		cur.Avatar = in.Avatar
	}
	if in.City != "" {
		cur.City = in.City
	}
	if in.Phone != "" {
		cur.Phone = in.Phone
	}
	if in.Birthday != "" {
		cur.Birthday = in.Birthday
	}

	if err := h.svc.SaveProfile(r.Context(), cur); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(service.ToResp(cur))
}

// @Summary  Ссылка для загрузки аватара
// @Tags     files
// @Produce  json
// @Success  200 {object} avatarPutResp
// @Router   /api/v1/avatar/presign [get]
func (h *AuthHandler) PresignAvatar(w http.ResponseWriter, r *http.Request) {
	cls, ok := middleware.FromCtx(r.Context())
	if !ok {
		http.Error(w, "no claims", http.StatusUnauthorized)
		return
	}

	putURL, publicURL, err := h.svc.GetAvatarPutURL(r.Context(), cls.UserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]string{
		"put_url":    putURL,
		"public_url": publicURL,
	})
}

// @Summary      Обновить только avatar в профиле
// @Tags         profile
// @Accept       json
// @Produce      json
// @Param        payload body      avatarReq  true  "URL картинки"
// @Success      204
// @Failure      400  {string}  string "bad json"
// @Failure      500  {string}  string "internal error"
// @Router       /api/v1/profile/avatar [patch]
func (h *AuthHandler) UpdateAvatar(w http.ResponseWriter, r *http.Request) {
	cls, ok := middleware.FromCtx(r.Context())
	if !ok {
		http.Error(w, "no claims", http.StatusUnauthorized)
		return
	}

	var in struct{ Avatar string }
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	if in.Avatar == "" {
		http.Error(w, "avatar URL is required", http.StatusBadRequest)
		return
	}

	cur, err := h.svc.Profile(r.Context(), cls.UserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	cur.Avatar = in.Avatar

	if err := h.svc.SaveProfile(r.Context(), cur); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
