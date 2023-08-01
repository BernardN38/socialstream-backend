package handler

import (
	"encoding/json"
	"net/http"

	"github.com/BernardN38/flutter-backend/user_service/service"
	"github.com/go-chi/jwtauth/v5"
)

type Handler struct {
	UserService  *service.UserService
	TokenManager *jwtauth.JWTAuth
}

func NewHandler(userService *service.UserService, tokenManager *jwtauth.JWTAuth) *Handler {
	return &Handler{
		UserService:  userService,
		TokenManager: tokenManager,
	}
}
func (h *Handler) CheckHealth(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("user service up and running"))
}

func (h *Handler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.UserService.GetAllUsers(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = json.NewEncoder(w).Encode(users)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
