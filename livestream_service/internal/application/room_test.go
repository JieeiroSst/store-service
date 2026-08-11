package application

import (
	"context"
	"testing"

	"github.com/JIeeiroSst/livestream-service/internal/domain/model"
)

func createRoomInput(ownerUserID, title string) model.CreateRoomInput {
	return model.CreateRoomInput{OwnerUserID: ownerUserID, Title: title}
}

func TestCreateRoomGeneratesAStreamKey(t *testing.T) {
	u := NewRoomUsecase(newFakeRoomRepository())

	room, err := u.CreateRoom(context.Background(), createRoomInput("owner-1", "My Room"))
	if err != nil {
		t.Fatalf("CreateRoom() error = %v", err)
	}
	if room.OwnerUserID != "owner-1" {
		t.Errorf("OwnerUserID = %q, want %q", room.OwnerUserID, "owner-1")
	}
	if room.StreamKey == "" {
		t.Error("expected a generated stream key")
	}
}

func TestRegenerateStreamKeyRejectsNonOwner(t *testing.T) {
	rooms := newFakeRoomRepository()
	u := NewRoomUsecase(rooms)
	room, err := u.CreateRoom(context.Background(), createRoomInput("owner-1", "My Room"))
	if err != nil {
		t.Fatalf("setup CreateRoom() error = %v", err)
	}

	if _, err := u.RegenerateStreamKey(context.Background(), room.ID, "someone-else", false); err != ErrForbidden {
		t.Fatalf("RegenerateStreamKey() error = %v, want ErrForbidden", err)
	}
}

func TestRegenerateStreamKeyAllowsOwner(t *testing.T) {
	rooms := newFakeRoomRepository()
	u := NewRoomUsecase(rooms)
	room, err := u.CreateRoom(context.Background(), createRoomInput("owner-1", "My Room"))
	if err != nil {
		t.Fatalf("setup CreateRoom() error = %v", err)
	}

	newKey, err := u.RegenerateStreamKey(context.Background(), room.ID, "owner-1", false)
	if err != nil {
		t.Fatalf("RegenerateStreamKey() error = %v", err)
	}
	if newKey == room.StreamKey {
		t.Error("expected a different stream key after regeneration")
	}
}

func TestRegenerateStreamKeyAllowsAdmin(t *testing.T) {
	rooms := newFakeRoomRepository()
	u := NewRoomUsecase(rooms)
	room, err := u.CreateRoom(context.Background(), createRoomInput("owner-1", "My Room"))
	if err != nil {
		t.Fatalf("setup CreateRoom() error = %v", err)
	}

	if _, err := u.RegenerateStreamKey(context.Background(), room.ID, "admin-user", true); err != nil {
		t.Fatalf("RegenerateStreamKey() as admin, error = %v", err)
	}
}
