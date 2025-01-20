package repositories

import (
	"github.com/JIeeiroSst/room-service/internal/core/domain/models"
	"github.com/JIeeiroSst/room-service/internal/core/ports"
	"gorm.io/gorm"
)

type messageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) ports.MessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) Create(message *models.Message) error {
	return r.db.Create(message).Error
}

func (r *messageRepository) FindByRoomID(roomID uint) ([]models.Message, error) {
	var messages []models.Message
	err := r.db.Where("room_id = ?", roomID).Order("timestamp asc").Find(&messages).Error
	return messages, err
}
