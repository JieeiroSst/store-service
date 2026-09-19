package application

import (
	"context"
	"fmt"

	"github.com/JIeeiroSst/threads-service/internal/domain/model"
	"github.com/JIeeiroSst/threads-service/internal/domain/port"
	"github.com/google/uuid"
)

type likeService struct {
	likes    port.LikeRepository
	posts    port.PostRepository
	comments port.CommentRepository
}

func NewLikeService(likes port.LikeRepository, posts port.PostRepository, comments port.CommentRepository) port.LikeUsecase {
	return &likeService{likes: likes, posts: posts, comments: comments}
}

func (s *likeService) LikePost(ctx context.Context, userID, postID string) error {
	if _, err := s.posts.GetByID(ctx, postID); err != nil {
		return fmt.Errorf("get post: %w", err)
	}
	exists, err := s.likes.Exists(ctx, userID, &postID, nil)
	if err != nil {
		return err
	}
	if exists {
		return port.ErrAlreadyLiked
	}
	if err := s.likes.Create(ctx, &model.Like{ID: uuid.NewString(), UserID: userID, PostID: &postID}); err != nil {
		return err
	}
	return s.posts.IncrementLikeCount(ctx, postID, 1)
}

func (s *likeService) UnlikePost(ctx context.Context, userID, postID string) error {
	deleted, err := s.likes.Delete(ctx, userID, &postID, nil)
	if err != nil {
		return err
	}
	if !deleted {
		return nil
	}
	return s.posts.IncrementLikeCount(ctx, postID, -1)
}

func (s *likeService) LikeComment(ctx context.Context, userID, commentID string) error {
	if _, err := s.comments.GetByID(ctx, commentID); err != nil {
		return fmt.Errorf("get comment: %w", err)
	}
	exists, err := s.likes.Exists(ctx, userID, nil, &commentID)
	if err != nil {
		return err
	}
	if exists {
		return port.ErrAlreadyLiked
	}
	return s.likes.Create(ctx, &model.Like{ID: uuid.NewString(), UserID: userID, CommentID: &commentID})
}

func (s *likeService) UnlikeComment(ctx context.Context, userID, commentID string) error {
	_, err := s.likes.Delete(ctx, userID, nil, &commentID)
	return err
}
