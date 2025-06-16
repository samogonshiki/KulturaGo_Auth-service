package http

import (
	"encoding/json"
	"net/http"

	"kulturago/auth-service/internal/middleware"
	"kulturago/auth-service/internal/service"
	"kulturago/auth-service/internal/tokens"
)

type AuthHandler struct {
	svc *service.Service
	mgr *tokens.Manager
}

func NewAuthHandler(s *service.Service, m *tokens.Manager) *AuthHandler { return &AuthHandler{s, m} }

// SignUp godoc
// @Summary      Регистрация
// @Description  Создаёт учётную запись по e-mail/паролю
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        payload body  struct{Email string; Password string}  true  "ввод"
// @Success      200 {object} map[string]int64    "user_id"
// @Failure      409 {string} string              "already exists"
// @Router       /auth/signup [post]
func (h *AuthHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	var in struct{ Email, Password string }
	if json.NewDecoder(r.Body).Decode(&in) != nil {
		http.Error(w, "bad json", 400)
		return
	}
	u, err := h.svc.SignUp(r.Context(), in.Email, in.Password)
	if err != nil {
		http.Error(w, err.Error(), 409)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"user_id": u.ID})
}

// SignIn godoc
// @Summary      Логин
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        payload  body  struct{Email string; Password string}  true  "JSON input"
// @Success      200 {object} map[string]string "access/refresh"
// @Failure      401 {string} string            "invalid creds"
// @Router       /api/v1/auth/signin [post]
func (h *AuthHandler) SignIn(w http.ResponseWriter, r *http.Request) {
	var in struct{ Email, Password string }
	if json.NewDecoder(r.Body).Decode(&in) != nil {
		http.Error(w, "bad json", 400)
		return
	}
	access, refresh, err := h.svc.SignIn(r.Context(), in.Email, in.Password)
	if err != nil {
		http.Error(w, err.Error(), 401)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]string{
		"access_token": access, "refresh_token": refresh,
	})
}

// Me godoc
// @Summary      Профиль текущего пользователя
// @Tags         auth
// @Security     Bearer
// @Produce      json
// @Success      200 {object} map[string]int64
// @Router       /api/v1/me [get]
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	id, _ := middleware.FromCtx(r.Context())
	_ = json.NewEncoder(w).Encode(map[string]any{"user_id": id})
}
