package application

import (
	"context"
	"fmt"

	"github.com/JIeeiroSst/threads-service/internal/domain/model"
	"github.com/JIeeiroSst/threads-service/internal/domain/port"
	"github.com/google/uuid"
)

type bookmarkService struct {
	bookmarks port.BookmarkRepository
	posts     port.PostRepository
	users     port.UserClient
}

func NewBookmarkService(bookmarks port.BookmarkRepository, posts port.PostRepository, users port.UserClient) port.BookmarkUsecase {
	return &bookmarkService{bookmarks: bookmarks, posts: posts, users: users}
}

func (s *bookmarkService) Bookmark(ctx context.Context, userID, postID string) error {
	if _, err := s.posts.GetByID(ctx, postID); err != nil {
		return fmt.Errorf("get post: %w", err)
	}
	exists, err := s.bookmarks.Exists(ctx, userID, postID)
	if err != nil {
		return err
	}
	if exists {
		return port.ErrAlreadyBookmarked
	}
	return s.bookmarks.Create(ctx, &model.Bookmark{ID: uuid.NewString(), UserID: userID, PostID: postID})
}

func (s *bookmarkService) Unbookmark(ctx context.Context, userID, postID string) error {
	_, err := s.bookmarks.Delete(ctx, userID, postID)
	return err
}

func (s *bookmarkService) ListBookmarks(ctx context.Context, userID string, cursor string, limit int) ([]model.Post, string, error) {
	posts, next, err := s.bookmarks.ListPostsByUser(ctx, userID, cursor, limit)
	if err != nil {
		return nil, "", err
	}
	resolveAuthors(ctx, s.users, posts)
	return posts, next, nil
}
