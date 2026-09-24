package port

import (
	"context"

	"github.com/JIeeiroSst/room-service/internal/domain/model"
)

type RoomRepository interface {
	Create(ctx context.Context, name, owner string) (model.Room, error)
	FindByID(ctx context.Context, id uint) (model.Room, error)
	ListByMember(ctx context.Context, username string) ([]model.Room, error)
	Members(ctx context.Context, roomID uint) ([]model.Member, error)
	Member(ctx context.Context, roomID uint, username string) (model.Member, error)
	AddMember(ctx context.Context, m model.Member) error
	RemoveMember(ctx context.Context, roomID uint, username string) error
}

type MessageRepository interface {
	Create(ctx context.Context, m *model.Message) error
	List(ctx context.Context, roomID, beforeID uint, limit int) ([]model.Message, error)
}

type UserDirectory interface {
	Login(ctx context.Context, username, password string) (model.Session, error)
	Refresh(ctx context.Context, refreshToken string) (model.Session, error)
	Validate(ctx context.Context, token string) (model.User, error)
}

type EventPublisher interface {
	Publish(ctx context.Context, e model.Event) error
}
