package handlers

import (
	"encoding/json"
	"gobr/internal/blog/repo"
	"net/http"
)

type BlogHandler struct {
	repo *repo.PostsRepo
}

func NewBlogHandler(repo *repo.PostsRepo) *BlogHandler {
	return &BlogHandler{
		repo: repo,
	}
}


func (b *BlogHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	var req repo.Post
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	post:=&repo.Post{
		Title: req.Title,
		Content: req.Content,
	}
}