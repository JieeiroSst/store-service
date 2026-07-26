package application

import (
	"context"

	"github.com/JIeeiroSst/automatic-payment-service/common"
	"github.com/JIeeiroSst/automatic-payment-service/internal/domain/model"
	"github.com/JIeeiroSst/automatic-payment-service/internal/domain/port"
	"github.com/JIeeiroSst/utils/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type paymentMethodService struct {
	pms port.PaymentMethodRepository
}

func NewPaymentMethodService(pms port.PaymentMethodRepository) port.PaymentMethodUsecase {
	return &paymentMethodService{pms: pms}
}

func (s *paymentMethodService) AddPaymentMethod(ctx context.Context, req port.AddPaymentMethodRequest) (*model.PaymentMethod, error) {
	lg := logger.WithContext(ctx)

	if req.UserID == uuid.Nil || req.Provider == "" {
		return nil, common.ErrInvalidRequest
	}

	pm := &model.PaymentMethod{
		PaymentMethodID: uuid.New(),
		UserID:          req.UserID,
		Provider:        req.Provider,
		TokenID:         req.TokenID,
		LastFourDigits:  req.LastFourDigits,
		ExpiryDate:      req.ExpiryDate,
		IsDefault:       req.IsDefault,
	}

	if req.IsDefault {
		if err := s.pms.ClearDefaultByUser(ctx, req.UserID); err != nil {
			lg.Error("AddPaymentMethod: clear previous default", zap.Error(err))
			return nil, common.ErrDBFailed
		}
	}

	if err := s.pms.Create(ctx, pm); err != nil {
		lg.Error("AddPaymentMethod", zap.Error(err))
		return nil, common.ErrDBFailed
	}
	return pm, nil
}

func (s *paymentMethodService) ListPaymentMethods(ctx context.Context, userID uuid.UUID) ([]model.PaymentMethod, error) {
	pms, err := s.pms.ListByUser(ctx, userID)
	if err != nil {
		logger.WithContext(ctx).Error("ListPaymentMethods", zap.Error(err))
		return nil, common.ErrDBFailed
	}
	return pms, nil
}

func (s *paymentMethodService) DeletePaymentMethod(ctx context.Context, id uuid.UUID) error {
	if err := s.pms.Delete(ctx, id); err != nil {
		logger.WithContext(ctx).Error("DeletePaymentMethod", zap.Error(err))
		return common.ErrDBFailed
	}
	return nil
}

func (s *paymentMethodService) SetDefaultPaymentMethod(ctx context.Context, userID, id uuid.UUID) error {
	lg := logger.WithContext(ctx)

	if _, err := s.pms.GetByID(ctx, id); err != nil {
		return common.ErrNotFound
	}
	if err := s.pms.ClearDefaultByUser(ctx, userID); err != nil {
		lg.Error("SetDefaultPaymentMethod: clear previous default", zap.Error(err))
		return common.ErrDBFailed
	}
	if err := s.pms.SetDefault(ctx, id); err != nil {
		lg.Error("SetDefaultPaymentMethod", zap.Error(err))
		return common.ErrDBFailed
	}
	return nil
}
