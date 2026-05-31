package service

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "log"
    "time"

    "gobr/internal/blog/dto"
    "gobr/internal/blog/repo"

    "github.com/google/uuid"
    "github.com/redis/go-redis/v9"
)

type PostService struct {
    repo        *repo.PostsRepo
    redisClient *redis.Client
}

func NewPostService(postRepo *repo.PostsRepo, redisClient *redis.Client) *PostService {
    return &PostService{
        repo:        postRepo,
        redisClient: redisClient,
    }
}

func (s *PostService) CreatePost(ctx context.Context, title, content string, authorID uuid.UUID) (*repo.Post, error) {
    if title == "" || content == "" {
        return nil, errors.New("title and content cannot be empty")
    }
    post := &repo.Post{
        Title:    title,
        Content:  content,
        AuthorID: authorID,
    }
    created, err := s.repo.Create(ctx, post)
    if err != nil {
        return nil, err
    }
    return created, nil
}

func (s *PostService) GetPosts(ctx context.Context, limit, offset int) ([]*repo.Post, error) {

    if limit < 1 || limit > 100 {
        return nil, errors.New("limit must be between 1 and 100")
    }
    if offset < 0 {
        return nil, errors.New("offset must be >= 0")
    }

    var cacheKey string
    if offset == 0 {
        cacheKey = fmt.Sprintf("posts:firstpage:limit:%d", limit)
    }

    if offset == 0 {
        cached, err := s.redisClient.Get(ctx, cacheKey).Result()
        if err == nil {
            var posts []*repo.Post
            if err := json.Unmarshal([]byte(cached), &posts); err == nil {
                return posts, nil
            }
        }
    }

    params := dto.ListPostsParams{
        Limit:  limit,
        Offset: offset,
    }
    posts, err := s.repo.GetPosts(ctx, params)
    if err != nil {
        return nil, err
    }

    if offset == 0 && len(posts) > 0 {
        jsonData, err := json.Marshal(posts)
        if err == nil {
            if err := s.redisClient.Set(ctx, cacheKey, jsonData, 5*time.Minute).Err(); err != nil {
                log.Printf("redis set error: %v", err)
            }
        }
    }

    return posts, nil
}

func (s *PostService) GetPostById(ctx context.Context, postId string) (*repo.Post, error) {
    post, err := s.repo.GetByID(ctx, postId)
    if err != nil {
        return nil, err
    }
    return post, nil
}

func (s *PostService) UpdatePost(ctx context.Context, postId string, req dto.UpdatePostRequest) (*repo.Post, error) {

    updated, err := s.repo.Update(ctx, postId, req)
    if err != nil {
        return nil, err
    }
    return updated, nil
}

func (s *PostService) DeletePost(ctx context.Context, postID string) error {
    err := s.repo.Delete(ctx, postID)
    if err != nil {
        if errors.Is(err, repo.ErrPostNotFound) {
            return err
        }
        return fmt.Errorf("failed to delete post: %w", err)
    }
    return nil
}