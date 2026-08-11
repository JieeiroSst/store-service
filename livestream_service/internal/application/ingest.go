package application

import (
	"context"
	"fmt"

	"github.com/JIeeiroSst/livestream-service/config"
	"github.com/JIeeiroSst/livestream-service/internal/domain/model"
	"github.com/JIeeiroSst/livestream-service/internal/domain/port"
)

// ingestUsecase serves the viewer/edge-facing reads and pre-publish node
// assignment. It has no TranscodeRunner dependency, so it can run on the
// stateless edge role without pulling in ffmpeg-process management.
type ingestUsecase struct {
	rooms    port.RoomRepository
	streams  port.StreamRepository
	vods     port.VODRepository
	assigner *nodeAssigner
	cfg      *config.Config
}

func NewIngestUsecase(rooms port.RoomRepository, streams port.StreamRepository, vods port.VODRepository, nodes port.NodeRegistry, cfg *config.Config) port.IngestUsecase {
	return &ingestUsecase{rooms: rooms, streams: streams, vods: vods, assigner: newNodeAssigner(nodes), cfg: cfg}
}

func (u *ingestUsecase) RequestIngestEndpoint(ctx context.Context, roomID, callerUserID string, callerIsAdmin bool) (*model.IngestEndpoint, error) {
	room, err := u.rooms.GetByID(ctx, roomID)
	if err != nil {
		return nil, fmt.Errorf("get room: %w", err)
	}
	if room.OwnerUserID != callerUserID && !callerIsAdmin {
		return nil, ErrForbidden
	}

	node, err := u.assigner.assign(ctx, room.StreamKey)
	if err != nil {
		return nil, fmt.Errorf("assign node: %w", err)
	}

	return &model.IngestEndpoint{
		RTMPURL:   node.Addr,
		NodeID:    node.ID,
		StreamKey: room.StreamKey,
	}, nil
}

func (u *ingestUsecase) GetActiveStream(ctx context.Context, roomID string) (*model.Stream, error) {
	return u.streams.GetActiveByRoomID(ctx, roomID)
}

func (u *ingestUsecase) ListRecordings(ctx context.Context, roomID string) ([]model.Recording, error) {
	return u.vods.ListByRoom(ctx, roomID)
}
