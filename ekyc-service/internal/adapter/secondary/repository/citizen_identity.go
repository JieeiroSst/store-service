package repository

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/ekyc-service/internal/domain/model"
	"github.com/JIeeiroSst/ekyc-service/internal/domain/port"
	"gorm.io/gorm"
)

type citizenIdentityRepository struct {
	db *gorm.DB
}

func NewCitizenIdentityRepository(db *gorm.DB) port.CitizenIdentityRepository {
	return &citizenIdentityRepository{db: db}
}

func (r *citizenIdentityRepository) Create(ctx context.Context, identity *model.CitizenIdentity) error {
	return r.db.WithContext(ctx).Create(identity).Error
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
