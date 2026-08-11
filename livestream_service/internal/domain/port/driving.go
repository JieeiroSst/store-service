package port

import (
	"context"
	"time"

	"github.com/JIeeiroSst/livestream-service/internal/domain/model"
)

type RoomUsecase interface {
	CreateRoom(ctx context.Context, input model.CreateRoomInput) (*model.Room, error)
	GetRoom(ctx context.Context, roomID string) (*model.Room, error)
	ListRooms(ctx context.Context, live bool) ([]model.Room, error)
	// RegenerateStreamKey requires callerUserID to own the room, unless
	// callerIsAdmin.
	RegenerateStreamKey(ctx context.Context, roomID, callerUserID string, callerIsAdmin bool) (string, error)
}

type IngestUsecase interface {
	// RequestIngestEndpoint requires callerUserID to own the room, unless
	// callerIsAdmin.
	RequestIngestEndpoint(ctx context.Context, roomID, callerUserID string, callerIsAdmin bool) (*model.IngestEndpoint, error)
	GetActiveStream(ctx context.Context, roomID string) (*model.Stream, error)
	ListRecordings(ctx context.Context, roomID string) ([]model.Recording, error)
	// GetPlaybackURL returns a signed, time-limited URL for the room's
	// live output (if live) or its most recent VOD recording otherwise.
	GetPlaybackURL(ctx context.Context, roomID string) (*model.PlaybackInfo, error)
}

// ModerationUsecase covers room-owner and platform-admin actions: forcing
// a stream to stop, muting a user in chat, and deleting a room outright.
// Every method enforces the same ownership-or-admin rule.
type ModerationUsecase interface {
	ForceEndStream(ctx context.Context, roomID, callerUserID string, callerIsAdmin bool) error
	BanFromChat(ctx context.Context, roomID, targetUserID, callerUserID string, callerIsAdmin bool, duration time.Duration) error
	UnbanFromChat(ctx context.Context, roomID, targetUserID, callerUserID string, callerIsAdmin bool) error
	DeleteRoom(ctx context.Context, roomID, callerUserID string, callerIsAdmin bool) error
}

type PublishUsecase interface {
	HandleOnPublish(ctx context.Context, streamKey, nodeID string) error
	HandleOnUnpublish(ctx context.Context, streamKey string) error
}

type ViewerUsecase interface {
	Heartbeat(ctx context.Context, roomID, sessionID string) error
	GetViewerCount(ctx context.Context, roomID string) (int64, error)
}

type ChatUsecase interface {
	Publish(ctx context.Context, msg model.ChatMessage) error
	Subscribe(ctx context.Context, roomID string) (<-chan model.ChatMessage, func(), error)
}

type NodeSchedulerUsecase interface {
	Heartbeat(ctx context.Context) error
	AssignNode(ctx context.Context, streamKey string) (*model.TranscodeNode, error)
	ReleaseNode(ctx context.Context, streamKey string) error
	CheckStale(ctx context.Context) error
}
