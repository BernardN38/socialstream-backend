package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/BernardN38/flutter-backend/user_service/service"
	"github.com/go-chi/chi/v5"
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

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	userId := chi.URLParam(r, "userId")
	userIdInt, err := strconv.Atoi(userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	user, err := h.UserService.GetUser(r.Context(), int32(userIdInt))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	err = json.NewEncoder(w).Encode(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func (h *Handler) UploadUserProfileImage(w http.ResponseWriter, r *http.Request) {
	userId := chi.URLParam(r, "userId")
	userIdInt, err := strconv.Atoi(userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Parse the incoming form data
	err = r.ParseMultipartForm(10 << 20) // 20MB limit
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	// Get the file from the "image" field in the form
	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Unable to get file from request", http.StatusBadRequest)
		return
	}
	defer file.Close()
	err = h.UserService.UpdateUserProfileImage(r.Context(), int32(userIdInt), file, header)
	if err != nil {
		log.Println(err)
		http.Error(w, "Unable to proccess upload", http.StatusBadRequest)
		return
	}
}
