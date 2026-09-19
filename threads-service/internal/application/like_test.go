package application

import (
	"context"
	"testing"

	"github.com/JIeeiroSst/threads-service/internal/domain/model"
	"github.com/JIeeiroSst/threads-service/internal/domain/port"
)

func TestLikePostRejectsDuplicateLikeAndCountsOnce(t *testing.T) {
	posts := newFakePostRepository()
	posts.posts["post-1"] = &model.Post{ID: "post-1", UserID: "owner-1"}
	u := NewLikeService(newFakeLikeRepository(), posts, newFakeCommentRepository())

	if err := u.LikePost(context.Background(), "user-1", "post-1"); err != nil {
		t.Fatalf("first LikePost() error = %v", err)
	}
	if err := u.LikePost(context.Background(), "user-1", "post-1"); err != port.ErrAlreadyLiked {
		t.Fatalf("second LikePost() error = %v, want ErrAlreadyLiked", err)
	}
	if got := posts.posts["post-1"].LikeCount; got != 1 {
		t.Errorf("LikeCount = %d, want 1", got)
	}
}

func TestUnlikePostIsANoOpWhenNotLiked(t *testing.T) {
	posts := newFakePostRepository()
	posts.posts["post-1"] = &model.Post{ID: "post-1", UserID: "owner-1"}
	u := NewLikeService(newFakeLikeRepository(), posts, newFakeCommentRepository())

	if err := u.UnlikePost(context.Background(), "user-1", "post-1"); err != nil {
		t.Fatalf("UnlikePost() error = %v", err)
	}
	if got := posts.posts["post-1"].LikeCount; got != 0 {
		t.Errorf("LikeCount = %d, want 0 - unliking a non-existent like must not decrement", got)
	}
}

func TestLikeThenUnlikePostRestoresCount(t *testing.T) {
	posts := newFakePostRepository()
	posts.posts["post-1"] = &model.Post{ID: "post-1", UserID: "owner-1"}
	u := NewLikeService(newFakeLikeRepository(), posts, newFakeCommentRepository())

	if err := u.LikePost(context.Background(), "user-1", "post-1"); err != nil {
		t.Fatalf("LikePost() error = %v", err)
	}
	if err := u.UnlikePost(context.Background(), "user-1", "post-1"); err != nil {
		t.Fatalf("UnlikePost() error = %v", err)
	}
	if got := posts.posts["post-1"].LikeCount; got != 0 {
		t.Errorf("LikeCount = %d, want 0", got)
	}
}
