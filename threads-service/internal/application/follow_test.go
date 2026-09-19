package application

import (
	"context"
	"testing"

	"github.com/JIeeiroSst/threads-service/internal/domain/port"
)

func TestFollowRejectsSelfFollow(t *testing.T) {
	u := NewFollowService(newFakeFollowRepository(), newFakeUserClient())

	if err := u.Follow(context.Background(), "user-1", "user-1"); err != port.ErrCannotFollowSelf {
		t.Fatalf("Follow() error = %v, want ErrCannotFollowSelf", err)
	}
}

func TestFollowRejectsDuplicateFollow(t *testing.T) {
	u := NewFollowService(newFakeFollowRepository(), newFakeUserClient())

	if err := u.Follow(context.Background(), "user-1", "user-2"); err != nil {
		t.Fatalf("first Follow() error = %v", err)
	}
	if err := u.Follow(context.Background(), "user-1", "user-2"); err != port.ErrAlreadyFollowing {
		t.Fatalf("second Follow() error = %v, want ErrAlreadyFollowing", err)
	}
}

func TestUnfollowThenFollowAgainSucceeds(t *testing.T) {
	u := NewFollowService(newFakeFollowRepository(), newFakeUserClient())

	if err := u.Follow(context.Background(), "user-1", "user-2"); err != nil {
		t.Fatalf("Follow() error = %v", err)
	}
	if err := u.Unfollow(context.Background(), "user-1", "user-2"); err != nil {
		t.Fatalf("Unfollow() error = %v", err)
	}
	if err := u.Follow(context.Background(), "user-1", "user-2"); err != nil {
		t.Fatalf("re-Follow() error = %v", err)
	}
}
