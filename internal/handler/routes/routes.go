package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"

	hhttp "kulturago/auth-service/internal/handler/http"
	"kulturago/auth-service/internal/middleware"
	"kulturago/auth-service/internal/service"
	"kulturago/auth-service/internal/tokens"
)

func NewRouter(svc *service.Service, mgr *tokens.Manager) http.Handler {
	r := chi.NewRouter()
	r.Use(cors.AllowAll().Handler)

	ah := hhttp.NewAuthHandler(svc, mgr)

	r.Route("/api/v1", func(r chi.Router) {

		r.Post("/auth/signup", ah.SignUp)
		r.Post("/auth/signin", ah.SignIn)

		r.Route("/auth/oauth/{provider}", func(r chi.Router) {
			r.Get("/login", ah.BeginOAuth)
			r.Get("/callback", ah.OAuthCallback)
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(mgr))
			r.Get("/me", ah.Me)
		})
	})

	return r
}
