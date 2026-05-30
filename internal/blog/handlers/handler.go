package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"gobr/internal/blog/middleware"
	"gobr/internal/blog/repo"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type PostsHandler struct {
	repo        *repo.PostsRepo
	redisClient *redis.Client
}

func NewPostsHandler(repo *repo.PostsRepo, redisClient *redis.Client) *PostsHandler {
	return &PostsHandler{
		repo:        repo,
		redisClient: redisClient,
	}
}

func (b *PostsHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := middleware.GetUser(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	// Convert string to uuid.UUID
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, "invalid user ID format", http.StatusInternalServerError)
		return
	}
	var req repo.Req
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if req.Title == "" || req.Content == "" {
		http.Error(w, "title and content are required", http.StatusBadRequest)
		return
	}

	post := &repo.Post{
		Title:    req.Title,
		Content:  req.Content,
		AuthorID: userID,
	}

	// Save to database
	created, err := b.repo.Create(r.Context(), post)
	if err != nil {
		http.Error(w, "failed to create post", http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

func (b *PostsHandler) GetPosts(w http.ResponseWriter, r *http.Request) {
	limitStr := r.FormValue("limit")
	offsetStr := r.FormValue("offset")
	limit := 20
	if limitStr != "" {
		parsed, err := strconv.Atoi(limitStr)
		if err != nil {
			http.Error(w, "invalid limit parameter", http.StatusBadRequest)
			return
		}
		if parsed > 0 && parsed <= 100 {
			limit = parsed
		} else {
			http.Error(w, "limit must be between 1 and 100", http.StatusBadRequest)
			return
		}
	}
	offset := 0
	if offsetStr != "" {
		val, err := strconv.Atoi(offsetStr)
		if err != nil {
			http.Error(w, "invalid offset parameter", http.StatusBadRequest)
			return
		}
		offset = val
	}
	cacheKey := fmt.Sprintf("posts:page:%d:limit:%d", offset/limit+1, limit)
	if cached, err := b.redisClient.Get(r.Context(), cacheKey).Result(); err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache", "HIT")
		w.Write([]byte(cached))
		return
	}
	params := repo.ListPostsParams{
		Limit:  limit,
		Offset: offset,
	}
	posts, err := b.repo.GetPosts(r.Context(), params)
	if err != nil {
		http.Error(w, "failed to fetch posts", http.StatusInternalServerError)
		return
	}
	jsonData, err := json.Marshal(posts)
	if err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
	_ = b.redisClient.Set(r.Context(), cacheKey, jsonData, 5*time.Minute).Err()
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", "MISS")
	w.Write(jsonData)
}

func (b *PostsHandler) GetPostById(w http.ResponseWriter, r *http.Request) {
	postId := r.PathValue("id")
	post, err := b.repo.GetByID(r.Context(), postId)
	if err != nil {
		if errors.Is(err, repo.ErrPostNotFound) {
			http.Error(w, "post not found", http.StatusNotFound)
		} else {
			http.Error(w, "failed to fetch post", http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(post); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (b *PostsHandler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	postId := r.PathValue("id")
	var req repo.UpdatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	post, err := b.repo.Update(r.Context(), postId, req)
	if err != nil {
		if errors.Is(err, repo.ErrPostNotFound) {
			http.Error(w, "post not found", http.StatusNotFound)
		} else {
			http.Error(w, "failed to fetch post", http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(post); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (b *PostsHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	postId := r.PathValue("id")
	if err := b.repo.Delete(r.Context(), postId); err != nil {
		if errors.Is(err, repo.ErrPostNotFound) {
			http.Error(w, "post not found", http.StatusNotFound)
		} else {
			http.Error(w, "failed to delete post", http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
