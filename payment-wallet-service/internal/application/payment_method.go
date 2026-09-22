package application

import (
	"context"
	"fmt"

	"github.com/JIeeiroSst/payment-wallet-service/internal/domain/model"
	"github.com/JIeeiroSst/payment-wallet-service/internal/domain/port"
	"github.com/google/uuid"
)

type paymentMethodService struct {
	methods port.PaymentMethodRepository
}

func NewPaymentMethodService(methods port.PaymentMethodRepository) port.PaymentMethodUsecase {
	return &paymentMethodService{methods: methods}
}

func (s *paymentMethodService) AddPaymentMethod(ctx context.Context, userID string, methodType model.PaymentMethodType, provider, accountNumber string) (*model.PaymentMethod, error) {
	if !methodType.Valid() {
		return nil, port.ErrInvalidPaymentMethodType
	}

	existing, err := s.methods.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list existing payment methods: %w", err)
	}

	method := &model.PaymentMethod{
		PaymentMethodID: uuid.NewString(),
		UserID:          userID,
		Type:            methodType,
		Provider:        provider,
		AccountNumber:   accountNumber,
		IsDefault:       len(existing) == 0,
		IsActive:        true,
	}
	if err := s.methods.Create(ctx, method); err != nil {
		return nil, fmt.Errorf("create payment method: %w", err)
	}
	return method, nil
}

func (s *paymentMethodService) ListPaymentMethods(ctx context.Context, userID string) ([]model.PaymentMethod, error) {
	return s.methods.ListByUser(ctx, userID)
}

func (s *paymentMethodService) RemovePaymentMethod(ctx context.Context, id string) error {
	return s.methods.Delete(ctx, id)
}

func (s *paymentMethodService) SetDefaultPaymentMethod(ctx context.Context, id string) (*model.PaymentMethod, error) {
	return s.methods.SetDefault(ctx, id)
}
