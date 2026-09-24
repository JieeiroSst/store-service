package repository

import (
	"context"

	"github.com/JIeeiroSst/car-rental-service/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReviewRepo struct{ db *gorm.DB }

func (r *ReviewRepo) Create(ctx context.Context, rv *model.Review) error {
	return r.db.WithContext(ctx).Create(rv).Error
}

func (r *ReviewRepo) ExistsForRental(ctx context.Context, rentalID uuid.UUID) (bool, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&model.Review{}).Where("rental_id = ?", rentalID).Count(&n).Error
	return n > 0, err
}

func (r *ReviewRepo) ListByVehicle(ctx context.Context, vehicleID uuid.UUID, offset, limit int) ([]model.Review, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.Review{}).Where("vehicle_id = ?", vehicleID)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var out []model.Review
	if err := q.Order("created_at DESC, review_id").Offset(offset).Limit(limit).Find(&out).Error; err != nil {
		return nil, 0, err
	}
	ids := make([]uuid.UUID, 0, len(out))
	for _, rv := range out {
		ids = append(ids, rv.UserID)
	}
	if len(ids) > 0 {
		var users []model.User
		if err := r.db.WithContext(ctx).Find(&users, "user_id IN ?", ids).Error; err != nil {
			return nil, 0, err
		}
		byID := make(map[uuid.UUID]*model.User, len(users))
		for i := range users {
			byID[users[i].ID] = &users[i]
		}
		for i := range out {
			out[i].User = byID[out[i].UserID]
		}
	}
	return out, total, nil
}
