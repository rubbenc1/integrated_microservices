package handlers

import (
	"encoding/json"
	"gobr/internal/auth/dto"
	"gobr/internal/auth/repo"
	"log"
	"net/http"
	"time"

	"github.com/IBM/sarama"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	repo      *repo.AuthRepo
	jwtSecret []byte
	producer  sarama.SyncProducer
}

func NewAuthhandler(repo *repo.AuthRepo, jwtSecret string, producer sarama.SyncProducer) *AuthHandler {
	return &AuthHandler{
		repo:      repo,
		jwtSecret: []byte(jwtSecret),
		producer:  producer,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	user := &repo.Auth{
		UserName:     req.UserName,
		Email:        req.Email,
		PasswordHash: string(hashed),
	}
	id, err := h.repo.Create(r.Context(), user)
	if err != nil {
		log.Printf("Create error: %v", err)
		switch err {
		case repo.ErrDuplicateEmail:
			http.Error(w, "email already exists", http.StatusConflict)
		case repo.ErrDuplicateUserName:
			http.Error(w, "username already exists", http.StatusConflict)
		default:
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}

	event:=dto.UserCreatedEvent{
		ID:   	  id.String(),
		Email:    user.Email,
		UserName: user.UserName,
	}
	eventBytes, err := json.Marshal(event)
	if err == nil {
		msg:=&sarama.ProducerMessage{
			Topic: "user_created",
			Value: sarama.StringEncoder(eventBytes),
		}
		partition, offset, err:=h.producer.SendMessage(msg)
		if err != nil {
        	log.Printf("Failed to send Kafka message: %v", err)
		} else {
			log.Printf("Message sent to partition %d at offset %d", partition, offset)
		}
	}

	token, err := h.generateToken(id.String())
	if err != nil {
		http.Error(w, "could not generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(dto.AuthResponse{
		ID:    id.String(),
		Token: token,
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	user, err:=h.repo.GetByEmail(r.Context(),req.Email)
	if err != nil {
        if err == repo.ErrUserNotFound {
            http.Error(w, "invalid credentials", http.StatusUnauthorized)
            return
        }
        http.Error(w, "internal error", http.StatusInternalServerError)
        return
    }
	if user == nil || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))!=nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
        return
	}

	token, err := h.generateToken(user.ID.String())
	if err != nil {
		http.Error(w, "could not generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "token": token,
    })
}

func (h *AuthHandler) generateToken(id string) (string, error) {
	claims := jwt.MapClaims{
		"id": id,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
		"iat": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(h.jwtSecret)
}
