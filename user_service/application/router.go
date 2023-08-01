package application

import (
	"time"

	"github.com/BernardN38/flutter-backend/user_service/handler"
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

	r.Get("/api/v1/users/health", h.CheckHealth)
	r.Get("/api/v1/users/all", h.GetAllUsers)
	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(h.TokenManager))
		r.Use(jwtauth.Authenticator)
	})
	return r
}
