package main

import (
	"database/sql"
	"fmt"
	"gobr/internal/blog/config"
	grpcclient "gobr/internal/blog/grpc_client"
	"gobr/internal/blog/handlers"
	"gobr/internal/blog/middleware"
	"gobr/internal/blog/repo"
	"log"
	"net/http"
)

func main() {
	cfg := config.LoadConfig()
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.POSTGRES_USER,
		cfg.POSTGRES_PASSWORD,
		cfg.DBHost,
		cfg.DBPort,
		cfg.POSTGRES_DB,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(10)
	if err = db.Ping(); err != nil {
		log.Fatalf("failed to ping db: %v", err)
	}
	grpcClient,err:=grpcclient.NewAuthClient(cfg.GRPC_CLIENT_PORT)
	if err!=nil {
		log.Fatalf("failed to connect to grpc server: %v", err)
	}
	authMiddleware:=middleware.AuthMiddleWare(grpcClient)
	postsRepo:=repo.NewPostsRepo(db)
	postsHandler:=handlers.NewPostsHandler(postsRepo)
	mux:=http.NewServeMux()
	createPostHandler:=http.HandlerFunc(postsHandler.CreatePost)
	getPostsHandler:=http.HandlerFunc(postsHandler.GetPosts)
	getPostById:=http.HandlerFunc(postsHandler.GetPostById)
	updatePost:=http.HandlerFunc(postsHandler.UpdatePost)
	deletePost:=http.HandlerFunc(postsHandler.DeletePost)

	mux.Handle("GET /posts", getPostsHandler)
	mux.Handle("GET /posts/{id}", getPostById)

	mux.Handle("POST /posts",authMiddleware(createPostHandler))
	mux.Handle("PATCH /posts/{id}", authMiddleware(updatePost))
	mux.Handle("DELETE /posts/{id}", authMiddleware(deletePost))

	log.Printf("Blog service listening on port %s", cfg.BLOG_SERVER_PORT)
	if err := http.ListenAndServe(":"+cfg.BLOG_SERVER_PORT, mux); err != nil {
		log.Fatal(err)
	}
}