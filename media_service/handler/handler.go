package handler

import (
	"net/http"

	"github.com/BernardN38/flutter-backend/media_service/service"
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
