package repo

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type AuthRepo struct {
	DB *sql.DB
}

func NewAuthRepo(db *sql.DB) *AuthRepo {
	return &AuthRepo{DB: db}
}

const uniqueViolationCode = "23505"

var (
	ErrDuplicateEmail    = errors.New("user with this email already exists")
	ErrDuplicateUserName = errors.New("user with this username already exists")
	ErrUserNotFound      = errors.New("user not found")
)

func (a *AuthRepo) Create(ctx context.Context, user *Auth) (uuid.UUID, error) {
	query := `
			INSERT INTO auth (username, email, password_hash)
			VALUES ($1, $2, $3)
			RETURNING id
	`
	var id uuid.UUID
	err := a.DB.QueryRowContext(ctx, query, user.UserName, user.Email, user.PasswordHash).Scan(&id)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == uniqueViolationCode {
			switch pqErr.Constraint {
			case "auth_email_key":
				return uuid.Nil, ErrDuplicateEmail
			case "auth_username_key":
				return uuid.Nil, ErrDuplicateUserName
			}
		}
		return uuid.Nil, err
	}
	user.ID = id
	return id, nil
}

func (a *AuthRepo) GetByEmail(ctx context.Context, email string) (*Auth, error) {
	query := `
			SELECT id, username, email, password_hash
			FROM auth
			WHERE email = $1
	`
	var user Auth
	err := a.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.UserName,
		&user.Email,
		&user.PasswordHash,
	)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}
