package repository

import (
	"context"

	"github.com/JIeeiroSst/car-rental-service/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LocationRepo struct{ db *gorm.DB }

func (r *LocationRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Location, error) {
	var l model.Location
	if err := r.db.WithContext(ctx).First(&l, "location_id = ?", id).Error; err != nil {
		return nil, wrapNotFound(err)
	}
	return &l, nil
}

func (r *LocationRepo) List(ctx context.Context, city, state, country string, offset, limit int) ([]model.Location, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.Location{})
	if city != "" {
		q = q.Where("city ILIKE ?", city)
	}
	if state != "" {
		q = q.Where("state ILIKE ?", state)
	}
	if country != "" {
		q = q.Where("country ILIKE ?", country)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var out []model.Location
	err := q.Order("name, location_id").Offset(offset).Limit(limit).Find(&out).Error
	return out, total, err
}
