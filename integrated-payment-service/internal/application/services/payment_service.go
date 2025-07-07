package services

import (
	"context"
	"errors"
	"time"

	"github.com/JIeeiroSst/integrated-payment-service/internal/domain/entities"
	"github.com/JIeeiroSst/integrated-payment-service/internal/domain/interfaces"
	"github.com/JIeeiroSst/integrated-payment-service/pkg/logger"

	"github.com/google/uuid"
)

type PaymentService struct {
	paymentRepo interfaces.PaymentRepository
	processors  map[entities.PaymentMethod]interfaces.PaymentProcessor
	logger      logger.Logger
}

func NewPaymentService(
	paymentRepo interfaces.PaymentRepository,
	processors map[entities.PaymentMethod]interfaces.PaymentProcessor,
	logger logger.Logger,
) *PaymentService {
	return &PaymentService{
		paymentRepo: paymentRepo,
		processors:  processors,
		logger:      logger,
	}
}

func (s *PaymentService) CreatePayment(ctx context.Context, req *CreatePaymentRequest) (*interfaces.PaymentResponse, error) {
	processor, exists := s.processors[req.PaymentMethod]
	if !exists {
		return nil, errors.New("unsupported payment method")
	}

	payment := &entities.Payment{
		ID:            uuid.New().String(),
		UserID:        req.UserID,
		Amount:        req.Amount,
		Currency:      req.Currency,
		PaymentMethod: req.PaymentMethod,
		Status:        entities.PaymentStatusPending,
		Description:   req.Description,
		Metadata:      req.Metadata,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.paymentRepo.Create(ctx, payment); err != nil {
		return nil, err
	}

	response, err := processor.CreatePayment(ctx, payment)
	if err != nil {
		payment.Status = entities.PaymentStatusFailed
		s.paymentRepo.Update(ctx, payment)
		return nil, err
	}

	payment.ExternalID = response.TransactionID
	payment.Status = response.Status
	s.paymentRepo.Update(ctx, payment)

	return response, nil
}

func (s *PaymentService) ProcessPayment(ctx context.Context, paymentID string) (*interfaces.PaymentResponse, error) {
	payment, err := s.paymentRepo.GetByID(ctx, paymentID)
	if err != nil {
		return nil, err
	}

	processor, exists := s.processors[payment.PaymentMethod]
	if !exists {
		return nil, errors.New("unsupported payment method")
	}

	response, err := processor.ProcessPayment(ctx, paymentID)
	if err != nil {
		return nil, err
	}

	payment.Status = response.Status
	payment.UpdatedAt = time.Now()
	s.paymentRepo.Update(ctx, payment)

	return response, nil
}

func (s *PaymentService) RefundPayment(ctx context.Context, paymentID string, amount float64) (*interfaces.PaymentResponse, error) {
	payment, err := s.paymentRepo.GetByID(ctx, paymentID)
	if err != nil {
		return nil, err
	}

	if payment.Status != entities.PaymentStatusSuccess {
		return nil, errors.New("payment not eligible for refund")
	}

	processor, exists := s.processors[payment.PaymentMethod]
	if !exists {
		return nil, errors.New("unsupported payment method")
	}

	response, err := processor.RefundPayment(ctx, paymentID, amount)
	if err != nil {
		return nil, err
	}

	payment.Status = entities.PaymentStatusRefunded
	payment.UpdatedAt = time.Now()
	s.paymentRepo.Update(ctx, payment)

	return response, nil
}

func (s *PaymentService) GetPayment(ctx context.Context, paymentID string) (*entities.Payment, error) {
	payment, err := s.paymentRepo.GetByID(ctx, paymentID)
	if err != nil {
		return nil, err
	}
	return payment, nil
}

func (s *PaymentService) GetPaymentStatus(ctx context.Context, paymentID string) (entities.PaymentStatus, error) {
	payment, err := s.paymentRepo.GetByID(ctx, paymentID)
	if err != nil {
		return entities.PaymentStatusUnknown, err
	}
	return payment.Status, nil
}

func (s *PaymentService) HandleMoMoWebhook(ctx context.Context, req *MoMoWebhookRequest) error {
	return nil
}

func (s *PaymentService) HandleVNPayWebhook(ctx context.Context, req *VNPayWebhookRequest) error {
	return nil
}

func (s *PaymentService) HandleZaloPayWebhook(ctx context.Context, req *ZaloPayWebhookRequest) error {
	return nil
}

func (s *PaymentService) HandleStripeWebhook(ctx context.Context, req *StripeWebhookRequest) error {
	return nil
}

type CreatePaymentRequest struct {
	UserID        string                 `json:"user_id" validate:"required"`
	Amount        float64                `json:"amount" validate:"required,gt=0"`
	Currency      string                 `json:"currency" validate:"required"`
	PaymentMethod entities.PaymentMethod `json:"payment_method" validate:"required"`
	Description   string                 `json:"description"`
	Metadata      string                 `json:"metadata"`
}

type MoMoWebhookRequest struct {
}

type VNPayWebhookRequest struct {
}

type ZaloPayWebhookRequest struct {
}

type StripeWebhookRequest struct {
}
