package application

import (
	"time"

	"github.com/BernardN38/flutter-backend/authentication_service/handler"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
)

func SetupRouter(h *handler.Handler) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/api/v1/auth/health", h.CheckHealth)

	// for debug purposes
	r.Get("/api/v1/auth/users", h.DEBUG_GetAllUsers)

	r.Post("/api/v1/auth/register", h.CreateUser)
	r.Post("/api/v1/auth/login", h.LoginUser)

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(h.TokenManager))
		r.Use(jwtauth.Authenticator)

		r.Get("/api/v1/auth/admin", h.AdminCheck)
	})
	return r
}
