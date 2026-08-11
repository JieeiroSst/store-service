package application

import (
	"context"
	"fmt"
	"time"

	"github.com/JIeeiroSst/livestream-service/config"
	"github.com/JIeeiroSst/livestream-service/internal/domain/model"
	"github.com/JIeeiroSst/livestream-service/internal/domain/port"
	"github.com/google/uuid"
)

type publishUsecase struct {
	rooms    port.RoomRepository
	streams  port.StreamRepository
	vods     port.VODRepository
	runner   port.TranscodeRunner
	assigner *nodeAssigner
	cfg      *config.Config
}

func NewPublishUsecase(
	rooms port.RoomRepository,
	streams port.StreamRepository,
	vods port.VODRepository,
	runner port.TranscodeRunner,
	nodes port.NodeRegistry,
	cfg *config.Config,
) port.PublishUsecase {
	return &publishUsecase{
		rooms: rooms, streams: streams, vods: vods,
		runner: runner, assigner: newNodeAssigner(nodes), cfg: cfg,
	}
}

func (u *publishUsecase) HandleOnPublish(ctx context.Context, streamKey, nodeID string) error {
	room, err := u.rooms.GetByStreamKey(ctx, streamKey)
	if err != nil {
		return ErrStreamKeyNotFound
	}

	if err := u.assigner.claim(ctx, streamKey, nodeID); err != nil {
		return err
	}

	now := time.Now()
	stream := &model.Stream{
		ID:        uuid.NewString(),
		RoomID:    room.ID,
		NodeID:    nodeID,
		Status:    model.StreamStatusLive,
		StartedAt: &now,
		CreatedAt: now,
	}
	if err := u.streams.Create(ctx, stream); err != nil {
		return fmt.Errorf("create stream: %w", err)
	}
	if err := u.rooms.UpdateStatus(ctx, room.ID, model.RoomStatusLive); err != nil {
		return fmt.Errorf("update room status: %w", err)
	}

	rtmpInput := fmt.Sprintf("%s/%s", u.cfg.Node.LocalRTMP, streamKey)
	if err := u.runner.Start(ctx, streamKey, rtmpInput); err != nil {
		return fmt.Errorf("start transcode job: %w", err)
	}
	return nil
}

func (u *publishUsecase) HandleOnUnpublish(ctx context.Context, streamKey string) error {
	room, err := u.rooms.GetByStreamKey(ctx, streamKey)
	if err != nil {
		return ErrStreamKeyNotFound
	}

	if err := u.runner.Stop(ctx, streamKey); err != nil {
		return fmt.Errorf("stop transcode job: %w", err)
	}

	stream, err := u.streams.GetActiveByRoomID(ctx, room.ID)
	if err == nil && stream != nil {
		endedAt := time.Now()
		if err := u.streams.Complete(ctx, stream.ID, endedAt); err != nil {
			return fmt.Errorf("complete stream: %w", err)
		}

		duration := 0
		if stream.StartedAt != nil {
			duration = int(endedAt.Sub(*stream.StartedAt).Seconds())
		}
		rec := &model.Recording{
			ID:              uuid.NewString(),
			StreamID:        stream.ID,
			RoomID:          room.ID,
			ObjectKey:       fmt.Sprintf("%s/master.m3u8", streamKey),
			DurationSeconds: duration,
			CreatedAt:       endedAt,
		}
		if err := u.vods.Create(ctx, rec); err != nil {
			return fmt.Errorf("create vod record: %w", err)
		}
	}

	if err := u.rooms.UpdateStatus(ctx, room.ID, model.RoomStatusOffline); err != nil {
		return fmt.Errorf("update room status: %w", err)
	}
	if err := u.assigner.release(ctx, streamKey); err != nil {
		return fmt.Errorf("release node: %w", err)
	}
	return nil
}
