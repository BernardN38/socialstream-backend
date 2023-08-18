package application

import (
	"time"

	"github.com/BernardN38/flutter-backend/media_service/handler"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
)

func SetupRouter(h *handler.Handler, tm *jwtauth.JWTAuth) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/api/v1/media/health", h.CheckHealth)
	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(tm))
		r.Use(jwtauth.Authenticator)
	})
	return r
}
