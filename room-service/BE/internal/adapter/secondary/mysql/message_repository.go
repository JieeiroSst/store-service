package mysql

import (
	"context"
	"slices"

	"github.com/JIeeiroSst/room-service/internal/domain/model"
	"gorm.io/gorm"
)

type MessageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) *MessageRepository { return &MessageRepository{db: db} }

func (r *MessageRepository) Create(ctx context.Context, m *model.Message) error {
	row := messageRow{RoomID: m.RoomID, Username: m.Username, Content: m.Content, CreatedAt: m.CreatedAt}
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return err
	}
	m.ID = row.ID
	return nil
}

func (r *MessageRepository) List(ctx context.Context, roomID, beforeID uint, limit int) ([]model.Message, error) {
	q := r.db.WithContext(ctx).Where("room_id = ?", roomID)
	if beforeID > 0 {
		q = q.Where("id < ?", beforeID)
	}
	var rows []messageRow
	if err := q.Order("id DESC").Limit(limit).Find(&rows).Error; err != nil {
		return nil, err
	}
	slices.Reverse(rows)
	msgs := make([]model.Message, len(rows))
	for i, row := range rows {
		msgs[i] = row.toModel()
	}
	return msgs, nil
}
