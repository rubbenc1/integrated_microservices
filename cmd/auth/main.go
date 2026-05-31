package main

import (
	"context"
	"database/sql"
	"fmt"
	pb "gobr/internal/auth"
	"gobr/internal/auth/config"
	grpcserver "gobr/internal/auth/grpc_server"
	"gobr/internal/auth/handlers"
	"gobr/internal/auth/repo"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IBM/sarama"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
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

	//Kafka settings
	cfgKafka := sarama.NewConfig()
	cfgKafka.Producer.RequiredAcks = sarama.WaitForAll
	cfgKafka.Producer.Retry.Max = 5
	cfgKafka.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer([]string{cfg.KAFKA_BROKER}, cfgKafka)
	if err != nil {
		log.Panic(err)
	}
	defer func() {
		if err := producer.Close(); err != nil {
			log.Panic(err)
		}
	}()

	authRepo := repo.NewAuthRepo(db)
	authHandler := handlers.NewAuthhandler(authRepo, cfg.JWT_SECRET, producer)

	http.HandleFunc("/register", authHandler.Register)
	http.HandleFunc("/login", authHandler.Login)

	httpSrv := &http.Server{Addr: fmt.Sprintf(":%s", cfg.AUTH_SERVER_PORT), Handler: nil}
	go func() {
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.GRPC_SERVER_PORT))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcSrv := grpc.NewServer()
	pb.RegisterAuthServiceServer(grpcSrv, grpcserver.NewAuthServiceManager(cfg.JWT_SECRET))
	go func() {
		if err := grpcSrv.Serve(lis); err != nil {
			log.Fatal(err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Println("Shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	grpcSrv.GracefulStop()
	if err := httpSrv.Shutdown(ctx); err != nil {
		log.Printf("HTTP shutdown error: %v", err)
	}
}
