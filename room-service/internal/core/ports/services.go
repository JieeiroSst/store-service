package ports

import "github.com/JIeeiroSst/room-service/internal/core/domain/models"

type RoomService interface {
	CreateRoom(room *models.Room) error
	ListRooms() ([]models.Room, error)
	JoinRoom(id uint, password, username string) (string, error)
	GetMessages(roomID uint) ([]models.Message, error)
}

type AuthService interface {
	GenerateToken(roomID uint, username string) (string, error)
	ValidateToken(token string) (*models.TokenClaims, error)
}

type WebSocketService interface {
	HandleConnection(roomID uint, username string, send chan []byte)
	BroadcastToRoom(roomID uint, message []byte)
}
