package port

import (
	"context"
	"io"
	"time"

	"github.com/JIeeiroSst/livestream-service/internal/domain/model"
)

type RoomRepository interface {
	Create(ctx context.Context, room *model.Room) error
	GetByID(ctx context.Context, id string) (*model.Room, error)
	GetByStreamKey(ctx context.Context, streamKey string) (*model.Room, error)
	List(ctx context.Context, live bool) ([]model.Room, error)
	UpdateStatus(ctx context.Context, id string, status model.RoomStatus) error
	UpdateStreamKey(ctx context.Context, id, streamKey string) error
}

type StreamRepository interface {
	Create(ctx context.Context, stream *model.Stream) error
	GetActiveByRoomID(ctx context.Context, roomID string) (*model.Stream, error)
	Complete(ctx context.Context, streamID string, endedAt time.Time) error
}

type VODRepository interface {
	Create(ctx context.Context, rec *model.Recording) error
	ListByRoom(ctx context.Context, roomID string) ([]model.Recording, error)
}

// NodeRegistry is the Redis-backed multi-node scheduler store: heartbeat,
// capacity ranking, and stream -> node assignment.
type NodeRegistry interface {
	Heartbeat(ctx context.Context, node model.TranscodeNode, ttl time.Duration) error
	TopCandidates(ctx context.Context, n int) ([]string, error)
	GetNode(ctx context.Context, nodeID string) (*model.TranscodeNode, bool, error)
	ReserveCapacity(ctx context.Context, nodeID string) (bool, error)
	ReleaseCapacity(ctx context.Context, nodeID string) error
	GetAssignment(ctx context.Context, streamKey string) (string, bool, error)
	SetAssignment(ctx context.Context, streamKey, nodeID string, ttl time.Duration) error
	ClearAssignment(ctx context.Context, streamKey string) error
}

// ViewerCounter tracks online viewers per room as a sliding window: a
// session is counted as long as it has heartbeated within the configured
// window, so a crashed/closed player ages out on its own without an
// explicit "leave" signal (which real viewers - who never connect to SRS
// directly in this architecture - have no reliable way to send anyway).
type ViewerCounter interface {
	Heartbeat(ctx context.Context, roomID, sessionID string, window time.Duration) error
	Get(ctx context.Context, roomID string, window time.Duration) (int64, error)
}

type ChatBroadcaster interface {
	Publish(ctx context.Context, msg model.ChatMessage) error
	Subscribe(ctx context.Context, roomID string) (<-chan model.ChatMessage, func(), error)
}

// ObjectStorage is the S3/MinIO-compatible sink for HLS segments,
// playlists, and finalized VOD files.
type ObjectStorage interface {
	PutObject(ctx context.Context, key string, body io.Reader, size int64, contentType, cacheControl string) error
	PresignGet(ctx context.Context, key string, ttl time.Duration) (string, error)
}

// TranscodeRunner owns the FFmpeg process lifecycle for a single stream
// key: start, stop, restart, and liveness checks.
type TranscodeRunner interface {
	Start(ctx context.Context, streamKey, rtmpInput string) error
	Stop(ctx context.Context, streamKey string) error
	Restart(ctx context.Context, streamKey string) error
	IsRunning(streamKey string) bool
	ActiveStreamKeys() []string
	// StaleSince returns stream keys whose HLS output hasn't seen a new
	// segment written in at least threshold - i.e. the ffmpeg process is
	// alive but has stopped producing output ("clinically dead").
	StaleSince(threshold time.Duration) []string
}
