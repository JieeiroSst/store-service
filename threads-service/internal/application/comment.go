package application

import (
	"context"
	"fmt"

	"github.com/JIeeiroSst/threads-service/internal/domain/model"
	"github.com/JIeeiroSst/threads-service/internal/domain/port"
	"github.com/google/uuid"
)

type commentService struct {
	comments port.CommentRepository
	posts    port.PostRepository
	users    port.UserClient
}

func NewCommentService(comments port.CommentRepository, posts port.PostRepository, users port.UserClient) port.CommentUsecase {
	return &commentService{comments: comments, posts: posts, users: users}
}

func (s *commentService) CreateComment(ctx context.Context, input model.CreateCommentInput) (*model.Comment, error) {
	if _, err := s.posts.GetByID(ctx, input.PostID); err != nil {
		return nil, fmt.Errorf("get post: %w", err)
	}

	comment := &model.Comment{
		ID:              uuid.NewString(),
		PostID:          input.PostID,
		UserID:          input.UserID,
		ParentCommentID: input.ParentCommentID,
		Content:         input.Content,
		MediaURLs:       input.MediaURLs,
	}
	if err := s.comments.Create(ctx, comment); err != nil {
		return nil, fmt.Errorf("create comment: %w", err)
	}
	if err := s.posts.IncrementCommentCount(ctx, input.PostID, 1); err != nil {
		return nil, fmt.Errorf("increment comment count: %w", err)
	}

	if author, err := s.users.GetUser(ctx, comment.UserID); err == nil {
		comment.Author = author
	}
	return comment, nil
}

func (s *commentService) ListComments(ctx context.Context, postID string, cursor string, limit int) ([]model.Comment, string, error) {
	comments, next, err := s.comments.ListByPost(ctx, postID, cursor, limit)
	if err != nil {
		return nil, "", err
	}

	ids := make([]string, 0, len(comments))
	seen := make(map[string]bool, len(comments))
	for _, c := range comments {
		if !seen[c.UserID] {
			seen[c.UserID] = true
			ids = append(ids, c.UserID)
		}
	}
	if len(ids) > 0 {
		if authors, err := s.users.GetUsers(ctx, ids); err == nil {
			for i := range comments {
				comments[i].Author = authors[comments[i].UserID]
			}
		}
	}
	return comments, next, nil
}

func (s *commentService) DeleteComment(ctx context.Context, id, callerUserID string) error {
	comment, err := s.comments.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if comment.UserID != callerUserID {
		return port.ErrForbidden
	}
	if err := s.comments.Delete(ctx, id); err != nil {
		return err
	}
	return s.posts.IncrementCommentCount(ctx, comment.PostID, -1)
}
