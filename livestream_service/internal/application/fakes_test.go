package application

import (
	"context"
	"fmt"
	"time"

	"github.com/JIeeiroSst/livestream-service/internal/domain/model"
)

func seedRoom(rooms *fakeRoomRepository, id, streamKey string) {
	_ = rooms.Create(context.Background(), &model.Room{
		ID: id, StreamKey: streamKey, Status: model.RoomStatusOffline,
	})
}

// --- fakeRoomRepository ---

type fakeRoomRepository struct {
	byID        map[string]*model.Room
	byStreamKey map[string]*model.Room
}

func newFakeRoomRepository() *fakeRoomRepository {
	return &fakeRoomRepository{
		byID:        make(map[string]*model.Room),
		byStreamKey: make(map[string]*model.Room),
	}
}

func (f *fakeRoomRepository) Create(ctx context.Context, room *model.Room) error {
	cp := *room
	f.byID[room.ID] = &cp
	f.byStreamKey[room.StreamKey] = &cp
	return nil
}

func (f *fakeRoomRepository) GetByID(ctx context.Context, id string) (*model.Room, error) {
	r, ok := f.byID[id]
	if !ok {
		return nil, fmt.Errorf("room %q not found", id)
	}
	cp := *r
	return &cp, nil
}

func (f *fakeRoomRepository) GetByStreamKey(ctx context.Context, streamKey string) (*model.Room, error) {
	r, ok := f.byStreamKey[streamKey]
	if !ok {
		return nil, fmt.Errorf("stream key %q not found", streamKey)
	}
	cp := *r
	return &cp, nil
}

func (f *fakeRoomRepository) List(ctx context.Context, live bool) ([]model.Room, error) {
	var out []model.Room
	for _, r := range f.byID {
		if !live || r.Status == model.RoomStatusLive {
			out = append(out, *r)
		}
	}
	return out, nil
}

func (f *fakeRoomRepository) UpdateStatus(ctx context.Context, id string, status model.RoomStatus) error {
	r, ok := f.byID[id]
	if !ok {
		return fmt.Errorf("room %q not found", id)
	}
	r.Status = status
	f.byStreamKey[r.StreamKey].Status = status
	return nil
}

func (f *fakeRoomRepository) UpdateStreamKey(ctx context.Context, id, streamKey string) error {
	r, ok := f.byID[id]
	if !ok {
		return fmt.Errorf("room %q not found", id)
	}
	delete(f.byStreamKey, r.StreamKey)
	r.StreamKey = streamKey
	f.byStreamKey[streamKey] = r
	return nil
}

func (f *fakeRoomRepository) Delete(ctx context.Context, id string) error {
	r, ok := f.byID[id]
	if !ok {
		return fmt.Errorf("room %q not found", id)
	}
	delete(f.byStreamKey, r.StreamKey)
	delete(f.byID, id)
	return nil
}

// --- fakeStreamRepository ---

type fakeStreamRepository struct {
	byID       map[string]*model.Stream
	activeByRm map[string]string // roomID -> streamID
}

func newFakeStreamRepository() *fakeStreamRepository {
	return &fakeStreamRepository{
		byID:       make(map[string]*model.Stream),
		activeByRm: make(map[string]string),
	}
}

func (f *fakeStreamRepository) Create(ctx context.Context, stream *model.Stream) error {
	cp := *stream
	f.byID[stream.ID] = &cp
	f.activeByRm[stream.RoomID] = stream.ID
	return nil
}

func (f *fakeStreamRepository) GetActiveByRoomID(ctx context.Context, roomID string) (*model.Stream, error) {
	id, ok := f.activeByRm[roomID]
	if !ok {
		return nil, fmt.Errorf("no active stream for room %q", roomID)
	}
	cp := *f.byID[id]
	return &cp, nil
}

func (f *fakeStreamRepository) Complete(ctx context.Context, streamID string, endedAt time.Time) error {
	s, ok := f.byID[streamID]
	if !ok {
		return fmt.Errorf("stream %q not found", streamID)
	}
	s.Status = model.StreamStatusEnded
	s.EndedAt = &endedAt
	delete(f.activeByRm, s.RoomID)
	return nil
}

// --- fakeVODRepository ---

type fakeVODRepository struct {
	created []model.Recording
}

func (f *fakeVODRepository) Create(ctx context.Context, rec *model.Recording) error {
	f.created = append(f.created, *rec)
	return nil
}

func (f *fakeVODRepository) ListByRoom(ctx context.Context, roomID string) ([]model.Recording, error) {
	var out []model.Recording
	for _, r := range f.created {
		if r.RoomID == roomID {
			out = append(out, r)
		}
	}
	return out, nil
}

// --- fakeTranscodeRunner ---

type fakeTranscodeRunner struct {
	started   map[string]string // streamKey -> rtmpInput
	stopped   map[string]bool
	stale     []string
	restarted []string
}

func newFakeTranscodeRunner() *fakeTranscodeRunner {
	return &fakeTranscodeRunner{started: make(map[string]string), stopped: make(map[string]bool)}
}

func (f *fakeTranscodeRunner) Start(ctx context.Context, streamKey, rtmpInput string) error {
	f.started[streamKey] = rtmpInput
	return nil
}
func (f *fakeTranscodeRunner) Stop(ctx context.Context, streamKey string) error {
	f.stopped[streamKey] = true
	return nil
}
func (f *fakeTranscodeRunner) Restart(ctx context.Context, streamKey string) error {
	f.restarted = append(f.restarted, streamKey)
	return nil
}
func (f *fakeTranscodeRunner) IsRunning(streamKey string) bool {
	_, ok := f.started[streamKey]
	return ok && !f.stopped[streamKey]
}
func (f *fakeTranscodeRunner) ActiveStreamKeys() []string {
	var out []string
	for k := range f.started {
		if !f.stopped[k] {
			out = append(out, k)
		}
	}
	return out
}
func (f *fakeTranscodeRunner) StaleSince(threshold time.Duration) []string { return f.stale }

// --- fakeNodeRegistry ---

type fakeNodeRegistry struct {
	nodes       map[string]*model.TranscodeNode
	assignments map[string]string
}

func newFakeNodeRegistry() *fakeNodeRegistry {
	return &fakeNodeRegistry{nodes: make(map[string]*model.TranscodeNode), assignments: make(map[string]string)}
}

func (f *fakeNodeRegistry) Heartbeat(ctx context.Context, node model.TranscodeNode, ttl time.Duration) error {
	cp := node
	f.nodes[node.ID] = &cp
	return nil
}

func (f *fakeNodeRegistry) TopCandidates(ctx context.Context, n int) ([]string, error) {
	var out []string
	for id := range f.nodes {
		out = append(out, id)
		if len(out) >= n {
			break
		}
	}
	return out, nil
}

func (f *fakeNodeRegistry) GetNode(ctx context.Context, nodeID string) (*model.TranscodeNode, bool, error) {
	n, ok := f.nodes[nodeID]
	if !ok {
		return nil, false, nil
	}
	cp := *n
	return &cp, true, nil
}

func (f *fakeNodeRegistry) ReserveCapacity(ctx context.Context, nodeID string) (bool, error) {
	n, ok := f.nodes[nodeID]
	if !ok || n.Capacity <= 0 {
		return false, nil
	}
	n.Capacity--
	return true, nil
}

func (f *fakeNodeRegistry) ReleaseCapacity(ctx context.Context, nodeID string) error {
	if n, ok := f.nodes[nodeID]; ok {
		n.Capacity++
	}
	return nil
}

func (f *fakeNodeRegistry) GetAssignment(ctx context.Context, streamKey string) (string, bool, error) {
	nodeID, ok := f.assignments[streamKey]
	return nodeID, ok, nil
}

func (f *fakeNodeRegistry) SetAssignment(ctx context.Context, streamKey, nodeID string, ttl time.Duration) error {
	f.assignments[streamKey] = nodeID
	return nil
}

func (f *fakeNodeRegistry) ClearAssignment(ctx context.Context, streamKey string) error {
	delete(f.assignments, streamKey)
	return nil
}

// --- fakeModerationStore ---

type fakeModerationStore struct {
	banned map[string]bool // "roomID:userID" -> banned
}

func newFakeModerationStore() *fakeModerationStore {
	return &fakeModerationStore{banned: make(map[string]bool)}
}

func (f *fakeModerationStore) Ban(ctx context.Context, roomID, userID string, ttl time.Duration) error {
	f.banned[roomID+":"+userID] = true
	return nil
}

func (f *fakeModerationStore) Unban(ctx context.Context, roomID, userID string) error {
	delete(f.banned, roomID+":"+userID)
	return nil
}

func (f *fakeModerationStore) IsBanned(ctx context.Context, roomID, userID string) (bool, error) {
	return f.banned[roomID+":"+userID], nil
}

// --- fakeNodeCaller ---

type fakeNodeCaller struct {
	forceUnpublished []string // nodeHTTPAddr values called
	err              error
}

func (f *fakeNodeCaller) ForceUnpublish(ctx context.Context, nodeHTTPAddr, streamKey string) error {
	if f.err != nil {
		return f.err
	}
	f.forceUnpublished = append(f.forceUnpublished, nodeHTTPAddr)
	return nil
}

// --- fakeChatBroadcaster ---

type fakeChatBroadcaster struct {
	published []model.ChatMessage
}

func (f *fakeChatBroadcaster) Publish(ctx context.Context, msg model.ChatMessage) error {
	f.published = append(f.published, msg)
	return nil
}

func (f *fakeChatBroadcaster) Subscribe(ctx context.Context, roomID string) (<-chan model.ChatMessage, func(), error) {
	ch := make(chan model.ChatMessage)
	return ch, func() {}, nil
}
