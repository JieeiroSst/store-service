package application

import (
	"context"

	"github.com/JIeeiroSst/livestream-service/internal/domain/model"
	"github.com/JIeeiroSst/livestream-service/internal/domain/port"
)

type chatUsecase struct {
	broadcaster port.ChatBroadcaster
}

func NewChatUsecase(broadcaster port.ChatBroadcaster) port.ChatUsecase {
	return &chatUsecase{broadcaster: broadcaster}
}

func (u *chatUsecase) Publish(ctx context.Context, msg model.ChatMessage) error {
	return u.broadcaster.Publish(ctx, msg)
}

func (u *chatUsecase) Subscribe(ctx context.Context, roomID string) (<-chan model.ChatMessage, func(), error) {
	return u.broadcaster.Subscribe(ctx, roomID)
}
