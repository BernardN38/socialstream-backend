package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/BernardN38/flutter-backend/service"
	"github.com/go-chi/jwtauth/v5"
	_ "github.com/lib/pq"
)

type Handler struct {
	AuthService  *service.AuthSerice
	TokenManager *jwtauth.JWTAuth
}

func NewHandler(authService *service.AuthSerice, tokenManager *jwtauth.JWTAuth) *Handler {
	return &Handler{
		AuthService:  authService,
		TokenManager: tokenManager,
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

func (h *Handler) LoginUser(w http.ResponseWriter, r *http.Request) {
	var loginUserInput LoginUserInput
	err := json.NewDecoder(r.Body).Decode(&loginUserInput)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = Validate(loginUserInput)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	userId, err := h.AuthService.LoginUser(r.Context(), service.LoginUserInput(loginUserInput))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	expirationTime := time.Now().Add(time.Minute * 30)
	_, tokenString, err := h.TokenManager.Encode(map[string]interface{}{"user_id": userId, "iss": "test", "exp": expirationTime})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Create an HttpOnly cookie to store the JWT token on the client-side
	cookie := &http.Cookie{
		Name:     "jwt",
		Value:    tokenString,
		Expires:  expirationTime, // Cookie expiration time (30 minutes)
		HttpOnly: true,           // HttpOnly flag for added security
		Secure:   false,
		Path:     "/",
	}

	http.SetCookie(w, cookie)

	// Respond with a success message
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"userId": userId,
	})
}

func (h *Handler) AdminCheck(w http.ResponseWriter, r *http.Request) {
	token, _, err := jwtauth.FromContext(r.Context())
	if err != nil {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	log.Printf("%+v", token)
	fmt.Fprintln(w, "Protected route - Access granted!")
}
