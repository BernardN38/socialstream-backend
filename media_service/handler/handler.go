package handler

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/BernardN38/socialstream-backend/media_service/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
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
func (h *Handler) GetAllMedia(w http.ResponseWriter, r *http.Request) {
	media, err := h.mediaService.GetAllMedia(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = json.NewEncoder(w).Encode(media)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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

func (h *Handler) UploadUserProfileImage(w http.ResponseWriter, r *http.Request) {
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
	// Parse the incoming form data
	err = r.ParseMultipartForm(10 << 20) // 20MB limit
	if err != nil {
		log.Println(err)
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	// Get the file from the "image" field in the form
	file, header, err := r.FormFile("image")
	if err != nil {
		log.Println(err)
		http.Error(w, "Unable to get file from request", http.StatusBadRequest)
		return
	}
	defer file.Close()
	err = h.mediaService.UploadUserProfileImage(r.Context(), int32(ctxUserId), file, header)
	if err != nil {
		log.Println(err)
		http.Error(w, "Unable to proccess upload", http.StatusBadRequest)
		return
	}
}

func (h *Handler) GetMedia(w http.ResponseWriter, r *http.Request) {
	mediaId := chi.URLParam(r, "mediaId")
	convertedMediaId, err := strconv.Atoi(mediaId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	object, err := h.mediaService.GetMediaCompressed(r.Context(), int32(convertedMediaId))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Set cache-related headers
	w.Header().Set("Cache-Control", "public, max-age=600") // Cache for 1 day (86400 seconds)
	w.Header().Set("Expires", time.Now().Add(time.Minute*10).Format(http.TimeFormat))
	w.Header().Set("ETag", uuid.NewString())
	_, err = io.Copy(w, object)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}
