package application

import (
	"context"
	"fmt"

	"github.com/JIeeiroSst/livestream-service/internal/domain/model"
	"github.com/JIeeiroSst/livestream-service/internal/domain/port"
)

type ingestUsecase struct {
	rooms    port.RoomRepository
	streams  port.StreamRepository
	vods     port.VODRepository
	assigner *nodeAssigner
}

func NewIngestUsecase(rooms port.RoomRepository, streams port.StreamRepository, vods port.VODRepository, nodes port.NodeRegistry) port.IngestUsecase {
	return &ingestUsecase{rooms: rooms, streams: streams, vods: vods, assigner: newNodeAssigner(nodes)}
}

func (u *ingestUsecase) RequestIngestEndpoint(ctx context.Context, roomID string) (*model.IngestEndpoint, error) {
	room, err := u.rooms.GetByID(ctx, roomID)
	if err != nil {
		return nil, fmt.Errorf("get room: %w", err)
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
