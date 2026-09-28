package repository

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/car-rental-service/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ProfileRepo struct{ db *gorm.DB }

func (r *ProfileRepo) Get(ctx context.Context, userID int64) (*model.CustomerProfile, error) {
	p := model.CustomerProfile{UserID: userID}
	err := r.db.WithContext(ctx).First(&p, "user_id = ?", userID).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	return &p, nil
}

func (r *ProfileRepo) Upsert(ctx context.Context, p *model.CustomerProfile) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"driving_license", "updated_at"}),
	}).Create(p).Error
}
