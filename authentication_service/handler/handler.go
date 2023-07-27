package handler

import (
	"encoding/json"
	"net/http"

	"github.com/BernardN38/flutter-backend/service"
	_ "github.com/lib/pq"
)

type Handler struct {
	AuthService *service.AuthSerice
}

func NewHandler(authService *service.AuthSerice) *Handler {
	return &Handler{
		AuthService: authService,
	}
}
func (h *Handler) CheckHealth(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("authentication service up and running"))
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var createUserInput CreateUserInput
	err := json.NewDecoder(r.Body).Decode(&createUserInput)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = Validate(createUserInput)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = h.AuthService.CreateUser(r.Context(), service.CreateUserInput{
		Username: createUserInput.Username,
		Email:    createUserInput.Email,
		Password: createUserInput.Password,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
}
