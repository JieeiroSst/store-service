package application

import (
	"context"
	"testing"
	"time"

	"github.com/JIeeiroSst/livestream-service/internal/domain/model"
)

func newTestModeration(t *testing.T) (*moderationUsecase, *fakeRoomRepository, *fakeStreamRepository, *fakeModerationStore, *fakeNodeRegistry, *fakeNodeCaller) {
	t.Helper()
	rooms := newFakeRoomRepository()
	streams := newFakeStreamRepository()
	moderation := newFakeModerationStore()
	nodes := newFakeNodeRegistry()
	nodeCaller := &fakeNodeCaller{}

	u := &moderationUsecase{rooms: rooms, streams: streams, moderation: moderation, nodes: nodes, nodeCaller: nodeCaller}
	return u, rooms, streams, moderation, nodes, nodeCaller
}

func TestBanFromChatRequiresOwnerOrAdmin(t *testing.T) {
	u, rooms, _, moderation, _, _ := newTestModeration(t)
	seedOwnedRoom(rooms, "room-1", "sk_test", "owner-1")

	if err := u.BanFromChat(context.Background(), "room-1", "troll", "someone-else", false, time.Hour); err != ErrForbidden {
		t.Fatalf("BanFromChat() by non-owner, error = %v, want ErrForbidden", err)
	}

	if err := u.BanFromChat(context.Background(), "room-1", "troll", "owner-1", false, time.Hour); err != nil {
		t.Fatalf("BanFromChat() by owner, error = %v", err)
	}
	banned, _ := moderation.IsBanned(context.Background(), "room-1", "troll")
	if !banned {
		t.Error("expected troll to be banned after BanFromChat")
	}
}

func TestUnbanFromChatRequiresOwnerOrAdmin(t *testing.T) {
	u, rooms, _, moderation, _, _ := newTestModeration(t)
	seedOwnedRoom(rooms, "room-1", "sk_test", "owner-1")
	_ = moderation.Ban(context.Background(), "room-1", "troll", time.Hour)

	if err := u.UnbanFromChat(context.Background(), "room-1", "troll", "someone-else", false); err != ErrForbidden {
		t.Fatalf("UnbanFromChat() by non-owner, error = %v, want ErrForbidden", err)
	}
	if err := u.UnbanFromChat(context.Background(), "room-1", "troll", "owner-1", false); err != nil {
		t.Fatalf("UnbanFromChat() by owner, error = %v", err)
	}
	banned, _ := moderation.IsBanned(context.Background(), "room-1", "troll")
	if banned {
		t.Error("expected troll to no longer be banned after UnbanFromChat")
	}
}

func TestForceEndStreamCallsTheAssignedNode(t *testing.T) {
	u, rooms, streams, _, nodes, nodeCaller := newTestModeration(t)
	seedOwnedRoom(rooms, "room-1", "sk_test", "owner-1")
	nodes.nodes["node-1"] = &model.TranscodeNode{ID: "node-1", HTTPAddr: "http://node-1:8080"}
	_ = streams.Create(context.Background(), &model.Stream{ID: "stream-1", RoomID: "room-1", NodeID: "node-1", Status: model.StreamStatusLive})

	if err := u.ForceEndStream(context.Background(), "room-1", "owner-1", false); err != nil {
		t.Fatalf("ForceEndStream() error = %v", err)
	}
	if len(nodeCaller.forceUnpublished) != 1 || nodeCaller.forceUnpublished[0] != "http://node-1:8080" {
		t.Errorf("expected ForceUnpublish called on http://node-1:8080, got %v", nodeCaller.forceUnpublished)
	}
}

func TestForceEndStreamRejectsNonOwner(t *testing.T) {
	u, rooms, streams, _, nodes, _ := newTestModeration(t)
	seedOwnedRoom(rooms, "room-1", "sk_test", "owner-1")
	nodes.nodes["node-1"] = &model.TranscodeNode{ID: "node-1", HTTPAddr: "http://node-1:8080"}
	_ = streams.Create(context.Background(), &model.Stream{ID: "stream-1", RoomID: "room-1", NodeID: "node-1", Status: model.StreamStatusLive})

	if err := u.ForceEndStream(context.Background(), "room-1", "someone-else", false); err != ErrForbidden {
		t.Fatalf("ForceEndStream() by non-owner, error = %v, want ErrForbidden", err)
	}
}

func TestDeleteRoomRejectsWhileLive(t *testing.T) {
	u, rooms, _, _, _, _ := newTestModeration(t)
	_ = rooms.Create(context.Background(), &model.Room{ID: "room-1", StreamKey: "sk_test", OwnerUserID: "owner-1", Status: model.RoomStatusLive})

	if err := u.DeleteRoom(context.Background(), "room-1", "owner-1", false); err == nil {
		t.Fatal("expected DeleteRoom to reject a live room")
	}
}

func TestDeleteRoomSucceedsWhenOffline(t *testing.T) {
	u, rooms, _, _, _, _ := newTestModeration(t)
	seedOwnedRoom(rooms, "room-1", "sk_test", "owner-1")

	if err := u.DeleteRoom(context.Background(), "room-1", "owner-1", false); err != nil {
		t.Fatalf("DeleteRoom() error = %v", err)
	}
	if _, err := rooms.GetByID(context.Background(), "room-1"); err == nil {
		t.Error("expected room to be deleted")
	}
}
