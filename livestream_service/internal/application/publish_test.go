package application

import (
	"context"
	"testing"

	"github.com/JIeeiroSst/livestream-service/config"
	"github.com/JIeeiroSst/livestream-service/internal/domain/model"
)

func newTestPublish(t *testing.T) (*publishUsecase, *fakeRoomRepository, *fakeStreamRepository, *fakeVODRepository, *fakeTranscodeRunner, *fakeNodeRegistry) {
	t.Helper()
	rooms := newFakeRoomRepository()
	streams := newFakeStreamRepository()
	vods := &fakeVODRepository{}
	runner := newFakeTranscodeRunner()
	nodes := newFakeNodeRegistry()
	nodes.nodes["node-1"] = &model.TranscodeNode{ID: "node-1", Addr: "rtmp://node-1:1935/live", Capacity: 5}

	cfg := &config.Config{Node: config.NodeConfig{LocalRTMP: "rtmp://127.0.0.1:1935/live"}}

	u := &publishUsecase{
		rooms: rooms, streams: streams, vods: vods, runner: runner,
		assigner: newNodeAssigner(nodes), cfg: cfg,
	}
	return u, rooms, streams, vods, runner, nodes
}

func TestHandleOnPublishStartsTranscodeAndClaimsNode(t *testing.T) {
	u, rooms, streams, _, runner, nodes := newTestPublish(t)
	seedRoom(rooms, "room-1", "sk_test")

	if err := u.HandleOnPublish(context.Background(), "sk_test", "node-1"); err != nil {
		t.Fatalf("HandleOnPublish() error = %v", err)
	}

	if rtmpInput, ok := runner.started["sk_test"]; !ok || rtmpInput != "rtmp://127.0.0.1:1935/live/sk_test" {
		t.Errorf("expected transcode job started with local rtmp input, got %q (started=%v)", rtmpInput, runner.started)
	}
	if assigned, ok, _ := nodes.GetAssignment(context.Background(), "sk_test"); !ok || assigned != "node-1" {
		t.Errorf("expected stream key assigned to node-1, got %q (ok=%v)", assigned, ok)
	}
	room, _ := rooms.GetByID(context.Background(), "room-1")
	if room.Status != model.RoomStatusLive {
		t.Errorf("expected room status live, got %q", room.Status)
	}
	if _, err := streams.GetActiveByRoomID(context.Background(), "room-1"); err != nil {
		t.Errorf("expected an active stream row, got error: %v", err)
	}
}

func TestHandleOnPublishRejectsUnknownStreamKey(t *testing.T) {
	u, _, _, _, _, _ := newTestPublish(t)

	err := u.HandleOnPublish(context.Background(), "sk_does_not_exist", "node-1")
	if err != ErrStreamKeyNotFound {
		t.Fatalf("HandleOnPublish() error = %v, want ErrStreamKeyNotFound", err)
	}
}

func TestHandleOnPublishRejectsWhenAssignedToAnotherNode(t *testing.T) {
	u, rooms, _, _, _, nodes := newTestPublish(t)
	seedRoom(rooms, "room-1", "sk_test")
	_ = nodes.SetAssignment(context.Background(), "sk_test", "node-2", 0)

	err := u.HandleOnPublish(context.Background(), "sk_test", "node-1")
	if err == nil {
		t.Fatal("expected an error when publishing to a node the stream key isn't assigned to")
	}
}

func TestHandleOnUnpublishStopsJobAndFinalizesVOD(t *testing.T) {
	u, rooms, _, vods, runner, _ := newTestPublish(t)
	seedRoom(rooms, "room-1", "sk_test")
	if err := u.HandleOnPublish(context.Background(), "sk_test", "node-1"); err != nil {
		t.Fatalf("setup HandleOnPublish() error = %v", err)
	}

	if err := u.HandleOnUnpublish(context.Background(), "sk_test"); err != nil {
		t.Fatalf("HandleOnUnpublish() error = %v", err)
	}

	if !runner.stopped["sk_test"] {
		t.Error("expected transcode job to be stopped")
	}
	room, _ := rooms.GetByID(context.Background(), "room-1")
	if room.Status != model.RoomStatusOffline {
		t.Errorf("expected room status offline, got %q", room.Status)
	}
	if len(vods.created) != 1 || vods.created[0].RoomID != "room-1" {
		t.Errorf("expected one VOD recording for room-1, got %+v", vods.created)
	}
}
