package repository

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/ekyc-service/internal/domain/model"
	"github.com/JIeeiroSst/ekyc-service/internal/domain/port"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type verificationRepository struct {
	db *gorm.DB
}

func NewVerificationRepository(db *gorm.DB) port.VerificationRepository {
	return &verificationRepository{db: db}
}

func (r *verificationRepository) Upsert(ctx context.Context, v *model.EkycVerification) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"status", "match_score", "verified_at", "updated_at"}),
	}).Create(v).Error
}

func (r *verificationRepository) GetByUserID(ctx context.Context, userID string) (*model.EkycVerification, error) {
	var v model.EkycVerification
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&v).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, port.ErrVerificationNotFound
	}
	if err != nil {
		return nil, err
	}
	return &v, nil
}
