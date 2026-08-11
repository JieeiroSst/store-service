package application

import (
	"context"
	"fmt"

	"github.com/JIeeiroSst/livestream-service/internal/domain/model"
	"github.com/JIeeiroSst/livestream-service/internal/domain/port"
	"github.com/JIeeiroSst/livestream-service/internal/infrastructure/metrics"
)

type chatUsecase struct {
	broadcaster port.ChatBroadcaster
	moderation  port.ModerationStore
}

func NewChatUsecase(broadcaster port.ChatBroadcaster, moderation port.ModerationStore) port.ChatUsecase {
	return &chatUsecase{broadcaster: broadcaster, moderation: moderation}
}

func (u *chatUsecase) Publish(ctx context.Context, msg model.ChatMessage) error {
	banned, err := u.moderation.IsBanned(ctx, msg.RoomID, msg.UserID)
	if err != nil {
		return fmt.Errorf("check chat ban: %w", err)
	}
	if banned {
		return ErrBanned
	}
	if err := u.broadcaster.Publish(ctx, msg); err != nil {
		return err
	}
	metrics.ChatMessagesTotal.Inc()
	return nil
}

func (u *chatUsecase) Subscribe(ctx context.Context, roomID string) (<-chan model.ChatMessage, func(), error) {
	return u.broadcaster.Subscribe(ctx, roomID)
}
