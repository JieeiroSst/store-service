package repository

import (
	"context"
	"errors"
	"time"

	"github.com/JIeeiroSst/payment-wallet-service/internal/domain/model"
	"github.com/JIeeiroSst/payment-wallet-service/internal/domain/port"
	"gorm.io/gorm"
)

type paymentRequestRepository struct {
	db *gorm.DB
}

func NewPaymentRequestRepository(db *gorm.DB) port.PaymentRequestRepository {
	return &paymentRequestRepository{db: db}
}

func (r *paymentRequestRepository) Create(ctx context.Context, request *model.PaymentRequest) error {
	request.CreatedAt = time.Now()
	return r.db.WithContext(ctx).Create(request).Error
}

func (r *paymentRequestRepository) GetByID(ctx context.Context, id string) (*model.PaymentRequest, error) {
	var request model.PaymentRequest
	if err := r.db.WithContext(ctx).Where("payment_request_id = ?", id).First(&request).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, port.ErrNotFound
		}
		return nil, err
	}
	return &request, nil
}

func (r *paymentRequestRepository) ListByWallet(ctx context.Context, walletID string) ([]model.PaymentRequest, error) {
	var requests []model.PaymentRequest
	err := r.db.WithContext(ctx).
		Where("requester_wallet_id = ? OR payer_wallet_id = ?", walletID, walletID).
		Order("created_at DESC").
		Find(&requests).Error
	return requests, err
}

func (r *paymentRequestRepository) UpdateStatus(ctx context.Context, id string, status model.PaymentRequestStatus, transferID *string) error {
	updates := map[string]interface{}{"status": status}
	if transferID != nil {
		updates["transfer_id"] = *transferID
	}
	res := r.db.WithContext(ctx).Model(&model.PaymentRequest{}).
		Where("payment_request_id = ? AND status = ?", id, model.PaymentRequestPending).
		Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return port.ErrPaymentRequestNotPending
	}
	return nil
}
