package application

import (
	"github.com/BernardN38/flutter-backend/handler"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

func SetupRouter(h *handler.Handler) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/api/v1/auth/health", h.CheckHealth)
	r.Post("/api/v1/auth/user", h.CreateUser)
	return r
}
