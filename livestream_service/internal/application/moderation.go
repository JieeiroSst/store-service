package application

import (
	"context"
	"fmt"
	"time"

	"github.com/JIeeiroSst/livestream-service/internal/domain/model"
	"github.com/JIeeiroSst/livestream-service/internal/domain/port"
)

type moderationUsecase struct {
	rooms      port.RoomRepository
	streams    port.StreamRepository
	moderation port.ModerationStore
	nodes      port.NodeRegistry
	nodeCaller port.NodeCaller
}

func NewModerationUsecase(
	rooms port.RoomRepository,
	streams port.StreamRepository,
	moderation port.ModerationStore,
	nodes port.NodeRegistry,
	nodeCaller port.NodeCaller,
) port.ModerationUsecase {
	return &moderationUsecase{rooms: rooms, streams: streams, moderation: moderation, nodes: nodes, nodeCaller: nodeCaller}
}

func (u *moderationUsecase) authorize(room *model.Room, callerUserID string, callerIsAdmin bool) error {
	if room.OwnerUserID != callerUserID && !callerIsAdmin {
		return ErrForbidden
	}
	return nil
}

// ForceEndStream stops a room's live stream from the edge role, which
// never runs ffmpeg itself: it looks up which node the active stream is
// assigned to and calls that node's internal force-unpublish route.
func (u *moderationUsecase) ForceEndStream(ctx context.Context, roomID, callerUserID string, callerIsAdmin bool) error {
	room, err := u.rooms.GetByID(ctx, roomID)
	if err != nil {
		return fmt.Errorf("get room: %w", err)
	}
	if err := u.authorize(room, callerUserID, callerIsAdmin); err != nil {
		return err
	}

	stream, err := u.streams.GetActiveByRoomID(ctx, roomID)
	if err != nil {
		return fmt.Errorf("room %q has no active stream: %w", roomID, err)
	}

	node, found, err := u.nodes.GetNode(ctx, stream.NodeID)
	if err != nil {
		return fmt.Errorf("get node: %w", err)
	}
	if !found || node.HTTPAddr == "" {
		return fmt.Errorf("transcode node %q is no longer reachable", stream.NodeID)
	}

	return u.nodeCaller.ForceUnpublish(ctx, node.HTTPAddr, room.StreamKey)
}

func (u *moderationUsecase) BanFromChat(ctx context.Context, roomID, targetUserID, callerUserID string, callerIsAdmin bool, duration time.Duration) error {
	room, err := u.rooms.GetByID(ctx, roomID)
	if err != nil {
		return fmt.Errorf("get room: %w", err)
	}
	if err := u.authorize(room, callerUserID, callerIsAdmin); err != nil {
		return err
	}
	return u.moderation.Ban(ctx, roomID, targetUserID, duration)
}

func (u *moderationUsecase) UnbanFromChat(ctx context.Context, roomID, targetUserID, callerUserID string, callerIsAdmin bool) error {
	room, err := u.rooms.GetByID(ctx, roomID)
	if err != nil {
		return fmt.Errorf("get room: %w", err)
	}
	if err := u.authorize(room, callerUserID, callerIsAdmin); err != nil {
		return err
	}
	return u.moderation.Unban(ctx, roomID, targetUserID)
}

func (u *moderationUsecase) DeleteRoom(ctx context.Context, roomID, callerUserID string, callerIsAdmin bool) error {
	room, err := u.rooms.GetByID(ctx, roomID)
	if err != nil {
		return fmt.Errorf("get room: %w", err)
	}
	if err := u.authorize(room, callerUserID, callerIsAdmin); err != nil {
		return err
	}
	if room.Status == model.RoomStatusLive {
		return fmt.Errorf("room %q is live; end the stream before deleting it", roomID)
	}
	return u.rooms.Delete(ctx, roomID)
}
