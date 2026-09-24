package application

import (
	"context"
	"log"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/JIeeiroSst/room-service/config"
	"github.com/JIeeiroSst/room-service/internal/domain/model"
	"github.com/JIeeiroSst/room-service/internal/domain/port"
)

type ChatService struct {
	rooms    port.RoomRepository
	messages port.MessageRepository
	events   port.EventPublisher
	maxLen   int
	maxLimit int
}

func NewChatService(rooms port.RoomRepository, messages port.MessageRepository, events port.EventPublisher, cfg *config.Config) *ChatService {
	return &ChatService{
		rooms:    rooms,
		messages: messages,
		events:   events,
		maxLen:   cfg.Chat.MaxMessageLen,
		maxLimit: cfg.Chat.HistoryLimit,
	}
}

func (s *ChatService) Authorize(ctx context.Context, actor model.User, roomID uint) error {
	_, err := requireMember(ctx, s.rooms, actor, roomID)
	return err
}

func (s *ChatService) Send(ctx context.Context, actor model.User, roomID uint, content string) (model.Message, error) {
	content = strings.TrimSpace(content)
	if content == "" || utf8.RuneCountInString(content) > s.maxLen {
		return model.Message{}, port.ErrInvalidInput
	}
	if err := s.Authorize(ctx, actor, roomID); err != nil {
		return model.Message{}, err
	}

	msg := model.Message{RoomID: roomID, Username: actor.Username, Content: content, CreatedAt: time.Now().UTC()}
	if err := s.messages.Create(ctx, &msg); err != nil {
		return model.Message{}, err
	}
	if err := s.events.Publish(ctx, model.Event{Type: model.EventMessage, RoomID: roomID, Message: &msg}); err != nil {
		log.Printf("publish message: %v", err)
	}
	return msg, nil
}

func (s *ChatService) History(ctx context.Context, actor model.User, roomID uint, beforeID uint, limit int) ([]model.Message, error) {
	if err := s.Authorize(ctx, actor, roomID); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > s.maxLimit {
		limit = s.maxLimit
	}
	return s.messages.List(ctx, roomID, beforeID, limit)
}
