package application

import (
	"time"

	"github.com/BernardN38/flutter-backend/user_service/handler"
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

	r.Get("/api/v1/users/health", h.CheckHealth)
	r.Get("/api/v1/users/all", h.GetAllUsers)
	r.Get("/api/v1/users/{userId}", h.GetUser)
	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(tm))
		r.Use(jwtauth.Authenticator)
		r.Post("/api/v1/users/{userId}/profileImage", h.UploadUserProfileImage)
	})
	return r
}
