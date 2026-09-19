package application

import (
	"context"
	"errors"
	"testing"

	"github.com/JIeeiroSst/threads-service/internal/domain/model"
	"github.com/JIeeiroSst/threads-service/internal/domain/port"
)

func TestBookmarkRejectsDuplicate(t *testing.T) {
	posts := newFakePostRepository()
	posts.posts["post-1"] = &model.Post{ID: "post-1", UserID: "owner-1"}
	u := NewBookmarkService(newFakeBookmarkRepository(), posts, newFakeUserClient())

	if err := u.Bookmark(context.Background(), "user-1", "post-1"); err != nil {
		t.Fatalf("first Bookmark() error = %v", err)
	}
	if err := u.Bookmark(context.Background(), "user-1", "post-1"); err != port.ErrAlreadyBookmarked {
		t.Fatalf("second Bookmark() error = %v, want ErrAlreadyBookmarked", err)
	}
}

func TestBookmarkRejectsMissingPost(t *testing.T) {
	u := NewBookmarkService(newFakeBookmarkRepository(), newFakePostRepository(), newFakeUserClient())

	if err := u.Bookmark(context.Background(), "user-1", "does-not-exist"); !errors.Is(err, port.ErrNotFound) {
		t.Fatalf("Bookmark() error = %v, want ErrNotFound", err)
	}
}

func TestUnbookmarkThenBookmarkAgainSucceeds(t *testing.T) {
	posts := newFakePostRepository()
	posts.posts["post-1"] = &model.Post{ID: "post-1", UserID: "owner-1"}
	u := NewBookmarkService(newFakeBookmarkRepository(), posts, newFakeUserClient())

	if err := u.Bookmark(context.Background(), "user-1", "post-1"); err != nil {
		t.Fatalf("Bookmark() error = %v", err)
	}
	if err := u.Unbookmark(context.Background(), "user-1", "post-1"); err != nil {
		t.Fatalf("Unbookmark() error = %v", err)
	}
	if err := u.Bookmark(context.Background(), "user-1", "post-1"); err != nil {
		t.Fatalf("re-Bookmark() error = %v", err)
	}
}
