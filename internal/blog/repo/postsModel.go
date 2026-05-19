package repo

import (
	"time"

	"github.com/google/uuid"
)

type Post struct {
    ID        uuid.UUID `json:"id"`
    Title     string    `json:"title"`
    Content   string    `json:"content"`
    AuthorID  uuid.UUID `json:"author_id"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}