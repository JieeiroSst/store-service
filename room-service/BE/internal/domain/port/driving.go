package port

import (
	"context"

	"github.com/JIeeiroSst/room-service/internal/domain/model"
)

type AuthService interface {
	Login(ctx context.Context, username, password string) (model.Session, error)
	Refresh(ctx context.Context, refreshToken string) (model.Session, error)
	Authenticate(ctx context.Context, token string) (model.User, error)
}

type RoomService interface {
	CreateRoom(ctx context.Context, actor model.User, name string) (model.Room, error)
	ListRooms(ctx context.Context, actor model.User) ([]model.Room, error)
	GetRoom(ctx context.Context, actor model.User, roomID uint) (model.Room, error)
	ListMembers(ctx context.Context, actor model.User, roomID uint) ([]model.Member, error)
	AddMember(ctx context.Context, actor model.User, roomID uint, username string) (model.Member, error)
	RemoveMember(ctx context.Context, actor model.User, roomID uint, username string) error
}

type ChatService interface {
	Authorize(ctx context.Context, actor model.User, roomID uint) error
	Send(ctx context.Context, actor model.User, roomID uint, content string) (model.Message, error)
	History(ctx context.Context, actor model.User, roomID uint, beforeID uint, limit int) ([]model.Message, error)
}
