package application

import (
	"time"

	"github.com/BernardN38/socialstream-backend/post_service/handler"
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

	r.Get("/api/v1/posts/health", h.CheckHealth)
	r.Get("/api/v1/posts/users/{userId}", h.GetPosts)

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(tm))
		r.Use(jwtauth.Authenticator)
		r.Get("/api/v1/posts/all", h.GetAllPosts)
		r.Post("/api/v1/posts", h.CreatePost)
		r.Delete("/api/v1/posts/{postId}", h.DeletePost)
	})
	return r
}
