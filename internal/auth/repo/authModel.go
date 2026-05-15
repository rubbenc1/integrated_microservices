package repo

import "github.com/google/uuid"

type Auth struct {
	ID           uuid.UUID `json:"id"`
	UserName     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
}

