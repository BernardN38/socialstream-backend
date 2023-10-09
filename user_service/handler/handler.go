package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/BernardN38/socialstream-backend/user_service/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
)

type Handler struct {
	UserService *service.UserService
}

func NewHandler(userService *service.UserService) *Handler {
	return &Handler{
		UserService: userService,
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

func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userId := chi.URLParam(r, "userId")
	userIdInt, err := strconv.Atoi(userId)
	if err != nil {

		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, claims, _ := jwtauth.FromContext(r.Context())
	ctxUserId := claims["user_id"].(float64)

	if int(ctxUserId) != userIdInt {
		log.Println("userId does not match token")
		http.Error(w, "unathorized", http.StatusUnauthorized)
		return
	}
	var updateUserReq UpdateUserRequest
	err = json.NewDecoder(r.Body).Decode(&updateUserReq)
	if err != nil {
		http.Error(w, "unable to decode json body", http.StatusBadRequest)
		return
	}
	err = Validate(updateUserReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = h.UserService.UpdateUser(r.Context(), int32(ctxUserId), service.UpdateUserInput(updateUserReq))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}
func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userId := chi.URLParam(r, "userId")
	userIdInt, err := strconv.Atoi(userId)
	if err != nil {

		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, claims, _ := jwtauth.FromContext(r.Context())
	ctxUserId := claims["user_id"].(float64)

	if int(ctxUserId) != userIdInt {
		log.Println("userId does not match token")
		http.Error(w, "unathorized", http.StatusUnauthorized)
		return
	}

	err = h.UserService.DeleteUser(r.Context(), int32(ctxUserId))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}
