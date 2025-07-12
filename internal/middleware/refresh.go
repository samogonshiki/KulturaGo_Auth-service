package middleware

import (
	"kulturago/auth-service/internal/util"
	"log"
	"net/http"
	"time"

	"kulturago/auth-service/internal/service"
	"kulturago/auth-service/internal/tokens"
)

func SlidingRefresh(svc *service.Service, mgr *tokens.Manager, threshold time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			accC, errAcc := r.Cookie("access_token")
			refC, errRef := r.Cookie("refresh_token")
			if errAcc != nil || errRef != nil || accC.Value == "" || refC.Value == "" {
				next.ServeHTTP(w, r)
				return
			}

			cls, err := mgr.Parse(accC.Value)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			remain := time.Until(cls.ExpiresAt.Time)
			needRenew := remain < threshold
			log.Printf("SlidingRefresh: until=%v, needRenew=%v", remain, needRenew)

			if needRenew {
				newAcc, newRef, err := svc.Refresh(r.Context(), refC.Value)
				if err == nil {
					util.Set(w, "access_token", newAcc,
						int(mgr.AccessTTLSeconds()), "/")
					util.Set(w, "refresh_token", newRef,
						int(mgr.RefreshTTLSeconds()), "/")

					r.Header.Set("Authorization", "Bearer "+newAcc)
				} else {
					log.Printf("SlidingRefresh: refresh failed: %v", err)
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
