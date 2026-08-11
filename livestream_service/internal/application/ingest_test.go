package application

import (
	"context"
	"testing"

	"github.com/JIeeiroSst/livestream-service/internal/domain/model"
)

func newTestIngest(t *testing.T) (*ingestUsecase, *fakeRoomRepository, *fakeStreamRepository, *fakeVODRepository, *fakeNodeRegistry) {
	t.Helper()
	rooms := newFakeRoomRepository()
	streams := newFakeStreamRepository()
	vods := &fakeVODRepository{}
	nodes := newFakeNodeRegistry()
	nodes.nodes["node-1"] = &model.TranscodeNode{ID: "node-1", Addr: "rtmp://node-1:1935/live", Capacity: 5}

	u := &ingestUsecase{rooms: rooms, streams: streams, vods: vods, assigner: newNodeAssigner(nodes)}
	return u, rooms, streams, vods, nodes
}

func TestRequestIngestEndpointAssignsANode(t *testing.T) {
	u, rooms, _, _, nodes := newTestIngest(t)
	seedRoom(rooms, "room-1", "sk_test")

	endpoint, err := u.RequestIngestEndpoint(context.Background(), "room-1")
	if err != nil {
		t.Fatalf("RequestIngestEndpoint() error = %v", err)
	}
	if endpoint.NodeID != "node-1" || endpoint.RTMPURL != "rtmp://node-1:1935/live" || endpoint.StreamKey != "sk_test" {
		t.Errorf("unexpected ingest endpoint: %+v", endpoint)
	}
	if assigned, ok, _ := nodes.GetAssignment(context.Background(), "sk_test"); !ok || assigned != "node-1" {
		t.Errorf("expected sk_test assigned to node-1, got %q (ok=%v)", assigned, ok)
	}
}

func TestRequestIngestEndpointErrorsForUnknownRoom(t *testing.T) {
	u, _, _, _, _ := newTestIngest(t)

	if _, err := u.RequestIngestEndpoint(context.Background(), "does-not-exist"); err == nil {
		t.Fatal("expected an error for an unknown room")
	}
}

func TestListRecordingsReturnsOnlyThatRoom(t *testing.T) {
	u, _, _, vods, _ := newTestIngest(t)
	vods.created = []model.Recording{
		{ID: "r1", RoomID: "room-1"},
		{ID: "r2", RoomID: "room-2"},
	}

	recs, err := u.ListRecordings(context.Background(), "room-1")
	if err != nil {
		t.Fatalf("ListRecordings() error = %v", err)
	}
	if len(recs) != 1 || recs[0].ID != "r1" {
		t.Errorf("ListRecordings() = %+v, want only room-1's recording", recs)
	}
}
