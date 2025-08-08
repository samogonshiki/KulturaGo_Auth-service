package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"kulturago/auth-service/internal/middleware"
)

func (h *AuthHandler) Security(w http.ResponseWriter, r *http.Request) {
	cls, ok := middleware.FromCtx(r.Context())
	if !ok {
		http.Error(w, "no claims", http.StatusUnauthorized)
		return
	}

	list, err := h.svc.Security(r.Context(), cls.UserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{"settings": list})
}

func (h *AuthHandler) ToggleSecurity(w http.ResponseWriter, r *http.Request) {
	cls, ok := middleware.FromCtx(r.Context())
	if !ok {
		http.Error(w, "no claims", http.StatusUnauthorized)
		return
	}

	key := chi.URLParam(r, "key")

	var in struct{ Enabled bool }
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	if err := h.svc.ToggleSecurity(r.Context(), cls.UserID, key, in.Enabled); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	cls, ok := middleware.FromCtx(r.Context())
	if !ok {
		http.Error(w, "no claims", http.StatusUnauthorized)
		return
	}

	var in struct{ Old, New string }
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	if err := h.svc.ChangePassword(r.Context(), cls.UserID, in.Old, in.New); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
