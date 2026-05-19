package repo

import (
	"context"
	"database/sql"
	"errors"
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

type ListPostsParams struct {
	Limit  int
	Offset int
}



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

func (p *PostsRepo) GetArticles(ctx context.Context, params ListPostsParams) ([]*Post, error) {

	if params.Limit <= 0 || params.Limit > 100 {
		params.Limit = 20
	}
	if params.Offset < 0 {
		params.Offset = 0
	}
	query := `
		SELECT *
		FROM posts
		LIMIT $1 OFFSET $2
	`
	rows, err := p.DB.QueryContext(ctx, query)
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

func (p *PostsRepo) UpdateArticle(ctx context.Context, id string) (*Post, error) {
	query:=`
			UPDATE posts
			SET Title =
	`
}
