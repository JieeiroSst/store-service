package repository

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/ekyc-service/internal/domain/model"
	"github.com/JIeeiroSst/ekyc-service/internal/domain/port"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type faceBiometricRepository struct {
	db *gorm.DB
}

func NewFaceBiometricRepository(db *gorm.DB) port.FaceBiometricRepository {
	return &faceBiometricRepository{db: db}
}

func (r *faceBiometricRepository) Upsert(ctx context.Context, face *model.FaceBiometric) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"template", "face_image_key", "quality_score", "updated_at"}),
	}).Create(face).Error
}

func (r *faceBiometricRepository) GetByUserID(ctx context.Context, userID string) (*model.FaceBiometric, error) {
	var face model.FaceBiometric
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&face).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, port.ErrFaceNotFound
	}
	if err != nil {
		return nil, err
	}
	return &face, nil
}
