package middleware

import (
	"context"
	"net/http"
	"strings"

	"kulturago/auth-service/internal/tokens"
)

type ctxKey int

const userIDKey ctxKey = 1

func FromCtx(ctx context.Context) (int64, bool) {
	v, ok := ctx.Value(userIDKey).(int64)
	return v, ok
}

func Auth(mgr *tokens.Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			var raw string

			if h := r.Header.Get("Authorization"); strings.HasPrefix(h, "Bearer ") {
				raw = strings.TrimPrefix(h, "Bearer ")
			}

			if raw == "" {
				if c, _ := r.Cookie("access_token"); c != nil {
					raw = c.Value
				}
			}

			if raw == "" {
				http.Error(w, "missing token", http.StatusUnauthorized)
				return
			}

			cls, err := mgr.Parse(raw)
			if err != nil {
				http.Error(w, "bad token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, cls.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
