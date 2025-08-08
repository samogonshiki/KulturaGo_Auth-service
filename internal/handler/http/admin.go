package http

import (
	"encoding/json"
	"errors"
	"kulturago/auth-service/internal/middleware"
	"net/http"

	"kulturago/auth-service/internal/service"
	"kulturago/auth-service/internal/tokens"
	"kulturago/auth-service/internal/util"
)

type AdminHandler struct {
	svc service.AdminService
	mgr *tokens.Manager
}

func NewAdminHandler(svc service.AdminService, mgr *tokens.Manager) *AdminHandler {
	return &AdminHandler{svc: svc, mgr: mgr}
}

type signInRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *AdminHandler) ASignIn(w http.ResponseWriter, r *http.Request) {
	var req signInRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorJSON(w, http.StatusBadRequest, err)
		return
	}

	tk, err := h.svc.ASignIn(r.Context(), req.Username, req.Password)
	if err != nil {
		errorJSON(w, http.StatusUnauthorized, err)
		return
	}

	util.Set(w, "access_token", tk.AccessToken, int(tk.ExpiresIn), "/")
	util.Set(w, "refresh_token", tk.RefreshToken, int(tk.RefreshExpiresIn), "/api/v1/admin")

	w.WriteHeader(http.StatusOK)
}

func (h *AdminHandler) ALogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("access_token")
	if err != nil {
		errorJSON(w, http.StatusUnauthorized, errors.New("no access cookie"))
		return
	}

	if err := h.svc.ALogOut(r.Context(), cookie.Value); err != nil {
		errorJSON(w, http.StatusInternalServerError, err)
		return
	}
	util.Clear(w, "access_token")
	util.ClearPath(w, "refresh_token", "/api/v1/admin")

	w.WriteHeader(http.StatusOK)
}

func (h *AdminHandler) Aaccess(w http.ResponseWriter, r *http.Request) {
	cls, ok := middleware.FromCtx(r.Context())
	if !ok {
		errorJSON(w, http.StatusUnauthorized, errors.New("no claims"))
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"authenticated": true,
		"admin_id":      cls.UserID,
	})
}

func errorJSON(w http.ResponseWriter, code int, err error) {
	writeJSON(w, code, map[string]string{"error": err.Error()})
}

func writeJSON(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}
