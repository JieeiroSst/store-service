package port

import (
	"context"

	"github.com/JIeeiroSst/livestream-service/internal/domain/model"
)

type RoomUsecase interface {
	CreateRoom(ctx context.Context, input model.CreateRoomInput) (*model.Room, error)
	GetRoom(ctx context.Context, roomID string) (*model.Room, error)
	ListRooms(ctx context.Context, live bool) ([]model.Room, error)
	RegenerateStreamKey(ctx context.Context, roomID string) (string, error)
}

type IngestUsecase interface {
	RequestIngestEndpoint(ctx context.Context, roomID string) (*model.IngestEndpoint, error)
	GetActiveStream(ctx context.Context, roomID string) (*model.Stream, error)
	ListRecordings(ctx context.Context, roomID string) ([]model.Recording, error)
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
