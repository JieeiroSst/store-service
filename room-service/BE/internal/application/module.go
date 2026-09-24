package application

import (
	"github.com/JIeeiroSst/room-service/internal/domain/port"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(NewAuthService, fx.As(new(port.AuthService))),
		fx.Annotate(NewRoomService, fx.As(new(port.RoomService))),
		fx.Annotate(NewChatService, fx.As(new(port.ChatService))),
	),
)
