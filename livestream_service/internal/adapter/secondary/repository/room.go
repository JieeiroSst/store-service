package repository

import (
	"context"
	"time"

	"github.com/JIeeiroSst/livestream-service/internal/domain/model"
	"github.com/JIeeiroSst/livestream-service/internal/domain/port"
	"gorm.io/gorm"
)

type roomRepository struct {
	db *gorm.DB
}

func NewRoomRepository(db *gorm.DB) port.RoomRepository {
	return &roomRepository{db: db}
}

func (r *roomRepository) Create(ctx context.Context, room *model.Room) error {
	return r.db.WithContext(ctx).Create(room).Error
}

func (r *roomRepository) GetByID(ctx context.Context, id string) (*model.Room, error) {
	var room model.Room
	if err := r.db.WithContext(ctx).First(&room, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &room, nil
}

func (r *roomRepository) GetByStreamKey(ctx context.Context, streamKey string) (*model.Room, error) {
	var room model.Room
	if err := r.db.WithContext(ctx).First(&room, "stream_key = ?", streamKey).Error; err != nil {
		return nil, err
	}
	return &room, nil
}

func (r *roomRepository) List(ctx context.Context, live bool) ([]model.Room, error) {
	q := r.db.WithContext(ctx).Order("created_at desc")
	if live {
		q = q.Where("status = ?", model.RoomStatusLive)
	}
	var rooms []model.Room
	if err := q.Find(&rooms).Error; err != nil {
		return nil, err
	}
	return rooms, nil
}

func (r *roomRepository) UpdateStatus(ctx context.Context, id string, status model.RoomStatus) error {
	return r.db.WithContext(ctx).Model(&model.Room{}).
		Where("id = ?", id).
		Updates(map[string]any{"status": status, "updated_at": time.Now()}).Error
}

func (r *roomRepository) UpdateStreamKey(ctx context.Context, id, streamKey string) error {
	return r.db.WithContext(ctx).Model(&model.Room{}).
		Where("id = ?", id).
		Updates(map[string]any{"stream_key": streamKey, "updated_at": time.Now()}).Error
}

func (r *roomRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&model.Room{}, "id = ?", id).Error
}
