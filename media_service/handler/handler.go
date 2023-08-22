package handler

import (
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/BernardN38/socialstream-backend/media_service/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	mediaService *service.MediaService
}

func New(m *service.MediaService) *Handler {
	return &Handler{
		mediaService: m,
	}
}

func (h *Handler) CheckHealth(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("media service is up and running"))
}

func (h *Handler) GetUserProfileImage(w http.ResponseWriter, r *http.Request) {
	userId := chi.URLParam(r, "userId")
	userIdInt, err := strconv.Atoi(userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	minioObject, err := h.mediaService.GetUserProfileImage(r.Context(), int32(userIdInt))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNoContent)
		return
	}

	// Set cache-related headers
	w.Header().Set("Cache-Control", "public, max-age=600") // Cache for 1 day (86400 seconds)
	w.Header().Set("Expires", time.Now().Add(time.Minute*10).Format(http.TimeFormat))
	w.Header().Set("ETag", uuid.NewString())
	_, err = io.Copy(w, minioObject)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
