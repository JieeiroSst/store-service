package application

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/JIeeiroSst/payment-service/internal/domain/model"
	"github.com/JIeeiroSst/payment-service/internal/domain/port"
)

const (
	defaultListLimit = 20
	maxListLimit     = 100
)

type paymentService struct {
	repo         port.PaymentRepository
	transactions port.TransactionRepository
	resolver     port.PaymentGatewayResolver
}

func NewPaymentService(repo port.PaymentRepository, transactions port.TransactionRepository, resolver port.PaymentGatewayResolver) port.PaymentUsecase {
	return &paymentService{repo: repo, transactions: transactions, resolver: resolver}
}

func (s *paymentService) CreatePayment(ctx context.Context, in port.CreatePaymentInput) (*model.Payment, error) {
	if !in.Provider.Valid() {
		return nil, fmt.Errorf("%w: %q", port.ErrInvalidProvider, in.Provider)
	}
	if in.Amount <= 0 {
		return nil, port.ErrInvalidAmount
	}
	if in.Currency == "" {
		return nil, port.ErrInvalidCurrency
	}

	if in.IdempotencyKey != "" {
		existing, err := s.repo.GetByIdempotencyKey(ctx, in.IdempotencyKey)
		if err == nil {
			return existing, nil
		}
		if !errors.Is(err, port.ErrNotFound) {
			return nil, fmt.Errorf("check idempotency key: %w", err)
		}
	}

	gateway, err := s.resolver.Resolve(in.Provider)
	if err != nil {
		return nil, err
	}

	payment := &model.Payment{
		Provider:       in.Provider,
		Amount:         in.Amount,
		Currency:       in.Currency,
		Status:         model.PaymentStatusPending,
		PayerEmail:     in.PayerEmail,
		Description:    in.Description,
		IdempotencyKey: in.IdempotencyKey,
	}
	created, err := s.repo.Create(ctx, payment)
	if err != nil {
		if in.IdempotencyKey != "" {
			if existing, lookupErr := s.repo.GetByIdempotencyKey(ctx, in.IdempotencyKey); lookupErr == nil {
				return existing, nil
			}
		}
		return nil, fmt.Errorf("persist payment: %w", err)
	}

	result, gatewayErr := gateway.CreatePayment(ctx, port.CreateGatewayPaymentInput{
		Amount:      in.Amount,
		Currency:    in.Currency,
		PayerEmail:  in.PayerEmail,
		Description: in.Description,
		ReferenceID: strconv.FormatInt(created.ID, 10),
		Metadata:    in.Metadata,
	})
	if gatewayErr != nil {
		failureReason := gatewayErr.Error()
		if err := s.repo.UpdateStatus(ctx, created.ID, model.PaymentStatusFailed, "", failureReason); err != nil {
			return nil, fmt.Errorf("record payment failure: %w", err)
		}
		created.Status = model.PaymentStatusFailed
		created.FailureReason = failureReason
		s.recordTransaction(ctx, created.ID, model.TransactionTypeCreate, model.PaymentStatusFailed, in.Amount, "", "", failureReason)
		return created, fmt.Errorf("%w: %s", port.ErrGatewayRequestFailed, gatewayErr)
	}

	if err := s.repo.UpdateStatus(ctx, created.ID, result.Status, result.ExternalID, ""); err != nil {
		return nil, fmt.Errorf("record payment result: %w", err)
	}
	created.Status = result.Status
	created.ExternalID = result.ExternalID
	created.UpdatedAt = time.Now()
	s.recordTransaction(ctx, created.ID, model.TransactionTypeCreate, result.Status, in.Amount, result.ExternalID, result.RawStatus, "")
	return created, nil
}

func (s *paymentService) GetPayment(ctx context.Context, id int64) (*model.Payment, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *paymentService) ListPayments(ctx context.Context, in port.ListPaymentsInput) (*port.PaymentList, error) {
	limit := in.Limit
	if limit <= 0 {
		limit = defaultListLimit
	}
	if limit > maxListLimit {
		limit = maxListLimit
	}
	offset := in.Offset
	if offset < 0 {
		offset = 0
	}

	items, total, err := s.repo.List(ctx, port.ListPaymentsFilter{
		Provider: in.Provider,
		Status:   in.Status,
		From:     in.From,
		To:       in.To,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		return nil, err
	}
	return &port.PaymentList{Items: items, Total: total}, nil
}

func (s *paymentService) RefundPayment(ctx context.Context, id int64, amount int64) (*model.Payment, error) {
	payment, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if payment.Status != model.PaymentStatusCaptured &&
		payment.Status != model.PaymentStatusAuthorized &&
		payment.Status != model.PaymentStatusPartiallyRefunded {
		return nil, port.ErrPaymentNotRefundable
	}

	remaining := payment.Amount - payment.RefundedAmount
	if amount == 0 {
		amount = remaining
	}
	if amount < 0 || amount > remaining {
		return nil, port.ErrInvalidRefundAmount
	}

	gateway, err := s.resolver.Resolve(payment.Provider)
	if err != nil {
		return nil, err
	}

	result, err := gateway.RefundPayment(ctx, port.RefundGatewayPaymentInput{
		ExternalID:         payment.ExternalID,
		Amount:             amount,
		Currency:           payment.Currency,
		OutstandingBalance: remaining,
	})
	if err != nil {
		s.recordTransaction(ctx, id, model.TransactionTypeRefund, payment.Status, amount, payment.ExternalID, "", err.Error())
		return nil, fmt.Errorf("%w: %s", port.ErrGatewayRequestFailed, err)
	}

	newRefundedAmount := payment.RefundedAmount + amount
	newStatus := model.PaymentStatusPartiallyRefunded
	if newRefundedAmount >= payment.Amount {
		newStatus = model.PaymentStatusRefunded
	}

	if err := s.repo.ApplyRefund(ctx, id, amount, newStatus); err != nil {
		return nil, fmt.Errorf("record refund: %w", err)
	}
	payment.RefundedAmount = newRefundedAmount
	payment.Status = newStatus
	payment.UpdatedAt = time.Now()
	s.recordTransaction(ctx, id, model.TransactionTypeRefund, newStatus, amount, payment.ExternalID, result.RawStatus, "")
	return payment, nil
}

func (s *paymentService) ListTransactions(ctx context.Context, paymentID int64) ([]model.Transaction, error) {
	if _, err := s.repo.GetByID(ctx, paymentID); err != nil {
		return nil, err
	}
	return s.transactions.ListByPaymentID(ctx, paymentID)
}

func (s *paymentService) HandleWebhook(ctx context.Context, provider model.Provider, payload []byte, headers map[string][]string) error {
	if !provider.Valid() {
		return fmt.Errorf("%w: %q", port.ErrInvalidProvider, provider)
	}

	gateway, err := s.resolver.Resolve(provider)
	if err != nil {
		return err
	}

	event, err := gateway.ParseWebhook(ctx, payload, headers)
	if err != nil {
		return fmt.Errorf("%w: %s", port.ErrInvalidWebhook, err)
	}

	payment, err := s.repo.GetByExternalID(ctx, event.ExternalID)
	if err != nil {
		return err
	}

	if err := s.repo.UpdateStatus(ctx, payment.ID, event.Status, event.ExternalID, ""); err != nil {
		return fmt.Errorf("record webhook update: %w", err)
	}
	s.recordTransaction(ctx, payment.ID, model.TransactionTypeWebhook, event.Status, payment.Amount, event.ExternalID, event.RawStatus, "")
	return nil
}


func (s *paymentService) recordTransaction(ctx context.Context, paymentID int64, txType model.TransactionType, status model.PaymentStatus, amount int64, externalID, rawStatus, errorMessage string) {
	_, err := s.transactions.Create(ctx, &model.Transaction{
		PaymentID:    paymentID,
		Type:         txType,
		Status:       status,
		Amount:       amount,
		ExternalID:   externalID,
		RawStatus:    rawStatus,
		ErrorMessage: errorMessage,
	})
	if err != nil {
		log.Printf("payment-service: record %s transaction for payment %d: %v", txType, paymentID, err)
	}
}
