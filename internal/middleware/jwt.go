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
			h := r.Header.Get("Authorization")
			if !strings.HasPrefix(h, "Bearer ") {
				http.Error(w, "missing bearer", 401)
				return
			}
			cls, err := mgr.Parse(strings.TrimPrefix(h, "Bearer "))
			if err != nil {
				http.Error(w, "bad token", 401)
				return
			}
			ctx := context.WithValue(r.Context(), userIDKey, cls.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
