package handler

import (
	"net/http"

	"github.com/BernardN38/flutter-backend/user_service/service"
	"github.com/go-chi/jwtauth/v5"
)

type Handler struct {
	UserService  *service.UserSerice
	TokenManager *jwtauth.JWTAuth
}

func NewHandler(userService *service.UserSerice, tokenManager *jwtauth.JWTAuth) *Handler {
	return &Handler{
		UserService:  userService,
		TokenManager: tokenManager,
	}
}
func (h *Handler) CheckHealth(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("user service up and running"))
}
