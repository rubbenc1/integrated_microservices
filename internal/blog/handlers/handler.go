package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"gobr/internal/blog/dto"
	"gobr/internal/blog/middleware"
	"gobr/internal/blog/repo"
	"gobr/internal/blog/service"

	"github.com/google/uuid"
)

type PostsHandler struct {
    postService *service.PostService
}

func NewPostsHandler(postService *service.PostService) *PostsHandler {
    return &PostsHandler{
        postService: postService,
    }
}

func (h *PostsHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
    userIDStr, ok := middleware.GetUser(r.Context())
    if !ok {
        http.Error(w, "unauthorized", http.StatusUnauthorized)
        return
    }
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, "invalid user ID format", http.StatusInternalServerError)
		return
	}
    var req struct {
        Title   string `json:"title"`
        Content string `json:"content"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid request", http.StatusBadRequest)
        return
    }

    created, err := h.postService.CreatePost(r.Context(), req.Title, req.Content, userID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(created)
}

func (h *PostsHandler) GetPosts(w http.ResponseWriter, r *http.Request) {
    limitStr := r.URL.Query().Get("limit")
    offsetStr := r.URL.Query().Get("offset")

    limit := 20
    if limitStr != "" {
        parsed, err := strconv.Atoi(limitStr)
        if err != nil {
            http.Error(w, "invalid limit parameter", http.StatusBadRequest)
            return
        }
        limit = parsed
    }

    offset := 0
    if offsetStr != "" {
        parsed, err := strconv.Atoi(offsetStr)
        if err != nil {
            http.Error(w, "invalid offset parameter", http.StatusBadRequest)
            return
        }
        offset = parsed
    }

    posts, err := h.postService.GetPosts(r.Context(), limit, offset)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(posts)
}

func (h *PostsHandler) GetPostById(w http.ResponseWriter, r *http.Request) {
    postId := r.PathValue("id")
    if postId == "" {
        http.Error(w, "missing post id", http.StatusBadRequest)
        return
    }

    post, err := h.postService.GetPostById(r.Context(), postId)
    if err != nil {
        if errors.Is(err, repo.ErrPostNotFound) {
            http.Error(w, "post not found", http.StatusNotFound)
        } else {
            http.Error(w, "failed to fetch post", http.StatusInternalServerError)
        }
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(post)
}

func (h *PostsHandler) UpdatePost(w http.ResponseWriter, r *http.Request) {
    postId := r.PathValue("id")
    if postId == "" {
        http.Error(w, "missing post id", http.StatusBadRequest)
        return
    }

    var req dto.UpdatePostRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid request", http.StatusBadRequest)
        return
    }

    updated, err := h.postService.UpdatePost(r.Context(), postId, req)
    if err != nil {
        if errors.Is(err, repo.ErrPostNotFound) {
            http.Error(w, "post not found", http.StatusNotFound)
        } else {
            http.Error(w, err.Error(), http.StatusBadRequest)
        }
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(updated)
}

func (h *PostsHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
    postId := r.PathValue("id")
    if postId == "" {
        http.Error(w, "missing post id", http.StatusBadRequest)
        return
    }

    err := h.postService.DeletePost(r.Context(), postId)
    if err != nil {
        if errors.Is(err, repo.ErrPostNotFound) {
            http.Error(w, "post not found", http.StatusNotFound)
        } else {
            http.Error(w, "failed to delete post", http.StatusInternalServerError)
        }
        return
    }

    w.WriteHeader(http.StatusNoContent)
}