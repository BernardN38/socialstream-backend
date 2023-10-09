package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/BernardN38/socialstream-backend/post_service/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
)

type Handler struct {
	postService *service.PostService
}

func NewHandler(postService *service.PostService) *Handler {
	return &Handler{
		postService: postService,
	}
}

func (h *Handler) CheckHealth(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("post server up and running"))
}

func (h *Handler) GetAllPosts(w http.ResponseWriter, r *http.Request) {
	posts, err := h.postService.GetAllPosts(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	err = json.NewEncoder(w).Encode(posts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
}
func (h *Handler) DeletePost(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	ctxUserId := claims["user_id"].(float64)

	postId := chi.URLParam(r, "postId")
	postIdInt, err := strconv.Atoi(postId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = h.postService.DeletePost(r.Context(), int32(postIdInt), int32(ctxUserId))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}
func (h *Handler) GetPosts(w http.ResponseWriter, r *http.Request) {
	userId := chi.URLParam(r, "userId")
	pageNo := r.URL.Query().Get("pageNo")
	pageSize := r.URL.Query().Get("pageSize")

	userIdInt, _ := strconv.Atoi(userId)
	pageNoInt, _ := strconv.Atoi(pageNo)
	pageSizeint, _ := strconv.Atoi(pageSize)
	if pageNoInt == 0 || pageSizeint == 0 || userIdInt <= 0 {
		http.Error(w, "invalid pagination or user id", http.StatusBadRequest)
		return
	}
	postPage, err := h.postService.GetUserPostsPaginated(r.Context(), int32(userIdInt), service.PostPageReq{
		PageNo:   int32(pageNoInt),
		PageSize: int32(pageSizeint),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	err = json.NewEncoder(w).Encode(postPage)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) CreatePost(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	ctxUserId := claims["user_id"].(float64)
	ctxUsername := claims["username"].(string)

	// Parse the incoming form data
	err := r.ParseMultipartForm(10 << 20) // 20MB limit
	if err != nil {
		log.Println(err)
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	var input service.CreatePostInput
	input.UserId = int32(ctxUserId)
	input.Body = r.FormValue("body")

	input.Username = ctxUsername
	file, header, err := r.FormFile("media")
	if err != nil && err != http.ErrMissingFile {
		log.Println("Error getting media file:", err)
		http.Error(w, "Unable to get file from request", http.StatusBadRequest)
		return
	}

	if err == nil && file != nil {
		defer file.Close()
		input.Media = file
		input.MediaType = header.Header.Get("Content-Type")
		input.MediaSize = header.Size
	}

	err = service.Validate(input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	err = h.postService.CreatePost(r.Context(), input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusCreated)
}
