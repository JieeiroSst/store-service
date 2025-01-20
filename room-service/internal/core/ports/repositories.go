package ports

import "github.com/JIeeiroSst/room-service/internal/core/domain/models"

type RoomRepository interface {
	Create(room *models.Room) error
	FindAll() ([]models.Room, error)
	FindByID(id uint) (*models.Room, error)
}

type MessageRepository interface {
	Create(message *models.Message) error
	FindByRoomID(roomID uint) ([]models.Message, error)
}
