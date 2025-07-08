package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"kulturago/auth-service/internal/handler/http"
	"kulturago/auth-service/internal/middleware"
	"kulturago/auth-service/internal/service"
	"kulturago/auth-service/internal/tokens"
	"time"
)

func NewRouter(svc *service.Service, mgr *tokens.Manager) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.SlidingRefresh(svc, mgr, 15*time.Minute))

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	ah := http.NewAuthHandler(svc, mgr)

	r.Route("/api/v1/auth", func(r chi.Router) {
		r.Post("/signup", ah.SignUp)
		r.Post("/signin", ah.SignIn)
		r.Post("/refresh", ah.Refresh)
		r.Post("/logout", ah.Logout)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(mgr))
		r.Get("/api/v1/me", ah.Me)
		r.Get("/api/v1/profile", ah.Profile)
		r.Put("/api/v1/profile", ah.SaveProfile)
		r.Get("/api/v1/avatar/presign", ah.PresignAvatar) //SCRUM-6
	})

	return r
}
