package repository

import (
	"context"

	"github.com/JIeeiroSst/livestream-service/internal/domain/model"
	"github.com/JIeeiroSst/livestream-service/internal/domain/port"
	"gorm.io/gorm"
)

type vodRepository struct {
	db *gorm.DB
}

func NewVODRepository(db *gorm.DB) port.VODRepository {
	return &vodRepository{db: db}
}

func (r *vodRepository) Create(ctx context.Context, rec *model.Recording) error {
	return r.db.WithContext(ctx).Create(rec).Error
}

func (r *vodRepository) ListByRoom(ctx context.Context, roomID string) ([]model.Recording, error) {
	var recs []model.Recording
	err := r.db.WithContext(ctx).
		Where("room_id = ?", roomID).
		Order("created_at desc").
		Find(&recs).Error
	if err != nil {
		return nil, err
	}
	return recs, nil
}
