package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"gobr/internal/blog/dto"
	"strings"

	"github.com/lib/pq"
)

type PostsRepo struct {
	DB *sql.DB
}

func NewPostsRepo(db *sql.DB) *PostsRepo {
	return &PostsRepo{
		DB: db,
	}
}

const notNullViolation = "23502"

var ErrPostNotFound = errors.New("post not found")

func (p *PostsRepo) Create(ctx context.Context, post *Post) (*Post, error) {
	query := `
			INSERT INTO posts (title, content, author_id)
			VALUES ($1, $2, $3)
			RETURNING id, title, content, created_at
	`
	var newPost Post
	err := p.DB.QueryRowContext(ctx, query, post.Title, post.Content, post.AuthorID).Scan(
		&newPost.ID,
		&newPost.Title,
		&newPost.Content,
		&newPost.CreatedAt,
	)
	if err != nil {
		var pqErr pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == notNullViolation {
			return nil, errors.New("missing required field")
		}
		return nil, err
	}
	return &newPost, nil
}

func (p *PostsRepo) GetPosts(ctx context.Context, params dto.ListPostsParams) ([]*Post, error) {
	query := `
		SELECT *
		FROM posts
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := p.DB.QueryContext(ctx, query, params.Limit, params.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var posts []*Post
	for rows.Next() {
		var p Post
		err := rows.Scan(
			&p.ID,
			&p.Title,
			&p.Content,
			&p.AuthorID,
			&p.CreatedAt,
			&p.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		posts = append(posts, &p)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return posts, nil
}

func (p *PostsRepo) Update(ctx context.Context, id string, req dto.UpdatePostRequest) (*Post, error) {
	if req.Title == nil && req.Content == nil {
		return nil, errors.New("no fields to update")
	}
	args := []interface{}{}
	setParts := []string{}
	idx := 1
	if req.Title != nil {
		setParts = append(setParts, fmt.Sprintf("title = $%d", idx))
		args = append(args, req.Title)
		idx++
	}
	if req.Content != nil {
		setParts = append(setParts, fmt.Sprintf("content = $%d", idx))
		args = append(args, req.Content)
		idx++
	}
	setParts = append(setParts, "updated_at = NOW()")
	args = append(args, id)
	query := fmt.Sprintf(`
        UPDATE posts
        SET %s
        WHERE id = $%d
        RETURNING id, title, content, author_id, created_at, updated_at
    `, strings.Join(setParts, ", "), idx)

	var updated Post
	err := p.DB.QueryRowContext(ctx, query, args...).Scan(
		&updated.ID,
		&updated.Title,
		&updated.Content,
		&updated.AuthorID,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPostNotFound
		}
		return nil, err
	}
	return &updated, nil
}

func (p *PostsRepo) GetByID(ctx context.Context, id string) (*Post, error) {
	query := `
        SELECT id, title, content, author_id, created_at, updated_at
        FROM posts
        WHERE id = $1
    `
	var post Post
	err := p.DB.QueryRowContext(ctx, query, id).Scan(
		&post.ID,
		&post.Title,
		&post.Content,
		&post.AuthorID,
		&post.CreatedAt,
		&post.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPostNotFound
		}
		return nil, err
	}
	return &post, nil
}

func (p *PostsRepo) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM posts WHERE id = $1`
	result, err := p.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrPostNotFound
	}
	return nil
}
