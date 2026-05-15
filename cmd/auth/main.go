package main

import (
	"database/sql"
	"fmt"
	"gobr/internal/auth/delivery/handlers"
	"gobr/internal/auth/repo"
	"gobr/internal/config"
	"log"
	"net/http"

	_ "github.com/lib/pq"
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
	authRepo := repo.NewAuthRepo(db)
	authHandler := handlers.NewAuthhandler(authRepo, cfg.JWT_TOKEN)
	http.HandleFunc("/register", authHandler.Register)
	http.HandleFunc("/login", authHandler.Login)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
