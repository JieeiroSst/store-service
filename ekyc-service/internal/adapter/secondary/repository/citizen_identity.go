package repository

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/ekyc-service/internal/domain/model"
	"github.com/JIeeiroSst/ekyc-service/internal/domain/port"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type citizenIdentityRepository struct {
	db *gorm.DB
}

func NewCitizenIdentityRepository(db *gorm.DB) port.CitizenIdentityRepository {
	return &citizenIdentityRepository{db: db}
}

func (r *citizenIdentityRepository) Create(ctx context.Context, identity *model.CitizenIdentity) error {
	db := r.db.WithContext(ctx)
	err := db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"source", "document_number", "surname", "given_names", "nationality", "date_of_birth", "sex",
			"date_of_expiry", "mrz_line1", "mrz_line2", "mrz_line3", "checksum_valid", "confidence",
			"front_image_key", "back_image_key", "nfc_verified", "chip_face_image_key", "updated_at",
		}),
	}).Create(identity).Error
	if err != nil {
		return err
	}
	return db.Where("user_id = ?", identity.UserID).First(identity).Error
}

func (r *citizenIdentityRepository) GetByUserID(ctx context.Context, userID string) (*model.CitizenIdentity, error) {
	var identity model.CitizenIdentity
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&identity).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, port.ErrIdentityNotFound
	}
	if err != nil {
		return nil, err
	}
	return &identity, nil
}
