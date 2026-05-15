package handlers

import (
	"encoding/json"
	"gobr/internal/auth/repo"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	repo      *repo.AuthRepo
	jwtSecret []byte
}

func NewAuthhandler(repo *repo.AuthRepo, jwtSecret string) *AuthHandler {
	return &AuthHandler{
		repo:      repo,
		jwtSecret: []byte(jwtSecret),
	}
}

type registerReq struct {
	UserName string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	ID    string `json:"id"`
	Token string `json:"token"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerReq
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
	token, err := h.generateToken(id.String())
	if err != nil {
		http.Error(w, "could not generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(authResponse{
		ID:    id.String(),
		Token: token,
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginReq
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
		"sub": id,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
		"iat": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(h.jwtSecret)

}
