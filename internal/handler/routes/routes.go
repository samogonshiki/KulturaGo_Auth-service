package routes

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"

	httpH "kulturago/auth-service/internal/handler/http"
	"kulturago/auth-service/internal/middleware"
	"kulturago/auth-service/internal/service"
	"kulturago/auth-service/internal/tokens"
)

func NewRouter(userSvc *service.Service, adminSvc service.AdminService, mgr *tokens.Manager) *chi.Mux {
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	r.Use(middleware.SlidingRefresh(userSvc, mgr, 15*time.Minute))

	uh := httpH.NewAuthHandler(userSvc, mgr)
	ah := httpH.NewAdminHandler(adminSvc, mgr)

	r.Route("/api/v1/auth", func(r chi.Router) {
		r.Post("/signup", uh.SignUp)
		r.Post("/signin", uh.SignIn)
		r.Post("/refresh", uh.Refresh)
		r.Post("/logout", uh.Logout)

		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(mgr))

			r.Get("/security", uh.Security)
			r.Patch("/security/{key}", uh.ToggleSecurity)
			r.Put("/security/password", uh.ChangePassword)

			r.Get("/me", uh.Me)
			r.Get("/profile", uh.Profile)
			r.Put("/profile", uh.SaveProfile)
			r.Patch("/profile/avatar", uh.UpdateAvatar)
			r.Get("/avatar/presign", uh.PresignAvatar)
		})
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(mgr))

			r.Get("/profile", uh.Profile)
			r.Put("/profile", uh.SaveProfile)
			r.Patch("/profile/avatar", uh.UpdateAvatar)

			r.Get("/security", uh.Security)
			r.Patch("/security/{key}", uh.ToggleSecurity)
			r.Put("/security/password", uh.ChangePassword)

			r.Get("/me", uh.Me)
			r.Get("/avatar/presign", uh.PresignAvatar)
		})
	})

	r.Route("/api/v1/admin", func(r chi.Router) {
		r.Post("/signin", ah.ASignIn)
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(mgr))
			r.Post("/logout", ah.ALogout)
			r.Get("/access", ah.Aaccess)
		})
	})

	return r
}
