package repository

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/payment-wallet-service/internal/domain/model"
	"github.com/JIeeiroSst/payment-wallet-service/internal/domain/port"
	"gorm.io/gorm"
)

type transferRepository struct {
	db *gorm.DB
}

func NewTransferRepository(db *gorm.DB) port.TransferRepository {
	return &transferRepository{db: db}
}

func (r *transferRepository) GetByID(ctx context.Context, transferID string) (*model.Transfer, error) {
	var transfer model.Transfer
	if err := r.db.WithContext(ctx).Where("transfer_id = ?", transferID).First(&transfer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, port.ErrNotFound
		}
		return nil, err
	}
	return &transfer, nil
}

func (r *transferRepository) GetByReferenceID(ctx context.Context, referenceID string) (*model.Transfer, error) {
	if referenceID == "" {
		return nil, port.ErrNotFound
	}
	var transfer model.Transfer
	if err := r.db.WithContext(ctx).Where("reference_id = ?", referenceID).First(&transfer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, port.ErrNotFound
		}
		return nil, err
	}
	return &transfer, nil
}
