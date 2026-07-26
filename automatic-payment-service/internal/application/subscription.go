package application

import (
	"context"
	"fmt"
	"time"

	"github.com/JIeeiroSst/automatic-payment-service/common"
	"github.com/JIeeiroSst/automatic-payment-service/internal/domain/model"
	"github.com/JIeeiroSst/automatic-payment-service/internal/domain/port"
	"github.com/JIeeiroSst/utils/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	MaxPaymentRetries = 3
	billingPeriod     = 30 * 24 * time.Hour
)

type subscriptionService struct {
	subs    port.SubscriptionRepository
	pms     port.PaymentMethodRepository
	txs     port.TransactionRepository
	inv     port.InvoiceRepository
	gateway port.PaymentGatewayPort
}

func NewSubscriptionService(
	subs port.SubscriptionRepository,
	pms port.PaymentMethodRepository,
	txs port.TransactionRepository,
	inv port.InvoiceRepository,
	gateway port.PaymentGatewayPort,
) port.SubscriptionUsecase {
	return &subscriptionService{subs: subs, pms: pms, txs: txs, inv: inv, gateway: gateway}
}

func (s *subscriptionService) CreateSubscription(ctx context.Context, req port.CreateSubscriptionRequest) (*model.Subscription, error) {
	lg := logger.WithContext(ctx)

	if req.UserID == uuid.Nil || req.PlanID == uuid.Nil || req.Amount < 0 {
		return nil, common.ErrInvalidRequest
	}
	currency := req.Currency
	if currency == "" {
		currency = "USD"
	}

	now := time.Now()
	sub := &model.Subscription{
		SubscriptionID: uuid.New(),
		UserID:         req.UserID,
		PlanID:         req.PlanID,
		Amount:         req.Amount,
		Currency:       currency,
		StartDate:      now,
		AutoRenewal:    req.AutoRenewal,
	}

	if req.TrialDays > 0 {
		trialEnd := now.AddDate(0, 0, req.TrialDays)
		sub.Status = model.SubscriptionTrial
		sub.TrialEndDate = &trialEnd
		sub.NextBillingDate = trialEnd

		if err := s.subs.Create(ctx, sub); err != nil {
			lg.Error("CreateSubscription: persist trial subscription", zap.Error(err))
			return nil, common.ErrDBFailed
		}
		return sub, nil
	}

	sub.Status = model.SubscriptionActive
	sub.NextBillingDate = now.Add(billingPeriod)

	if req.Amount > 0 {
		pm, err := s.resolvePaymentMethod(ctx, req.UserID, req.PaymentMethodID)
		if err != nil {
			return nil, err
		}

		result, err := s.gateway.Charge(ctx, port.ChargeRequest{
			PaymentMethod: pm,
			Amount:        sub.Amount,
			Currency:      sub.Currency,
			Description:   fmt.Sprintf("Initial payment for subscription %s", sub.SubscriptionID),
		})
		if err != nil {
			lg.Error("CreateSubscription: gateway unreachable", zap.Error(err))
			return nil, common.ErrGatewayUnavailable
		}
		if !result.Success {
			lg.Info("CreateSubscription: initial charge declined", zap.String("userId", req.UserID.String()), zap.String("reason", result.ErrorMessage))
			return nil, common.ErrPaymentFailed
		}

		if err := s.subs.Create(ctx, sub); err != nil {
			lg.Error("CreateSubscription: persist subscription", zap.Error(err))
			return nil, common.ErrDBFailed
		}
		if err := s.recordSuccessfulCharge(ctx, sub, pm, result); err != nil {
			lg.Error("CreateSubscription: record initial charge", zap.Error(err))
		}
		return sub, nil
	}

	if err := s.subs.Create(ctx, sub); err != nil {
		lg.Error("CreateSubscription: persist subscription", zap.Error(err))
		return nil, common.ErrDBFailed
	}
	return sub, nil
}

func (s *subscriptionService) GetSubscription(ctx context.Context, id uuid.UUID) (*model.Subscription, error) {
	sub, err := s.subs.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return sub, nil
}

func (s *subscriptionService) CancelSubscription(ctx context.Context, id uuid.UUID) (*model.Subscription, error) {
	lg := logger.WithContext(ctx)

	sub, err := s.subs.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	sub.Status = model.SubscriptionCancelled
	sub.AutoRenewal = false
	sub.EndDate = &now

	if err := s.subs.Update(ctx, sub); err != nil {
		lg.Error("CancelSubscription", zap.Error(err))
		return nil, common.ErrDBFailed
	}
	return sub, nil
}

func (s *subscriptionService) ProcessDueRenewals(ctx context.Context) (int, error) {
	lg := logger.WithContext(ctx)

	due, err := s.subs.ListDueForRenewal(ctx, time.Now())
	if err != nil {
		lg.Error("ProcessDueRenewals: list due subscriptions", zap.Error(err))
		return 0, common.ErrDBFailed
	}

	for i := range due {
		s.renewOne(ctx, &due[i])
	}
	return len(due), nil
}

func (s *subscriptionService) renewOne(ctx context.Context, sub *model.Subscription) {
	lg := logger.WithContext(ctx)

	pm, err := s.pms.GetDefaultByUser(ctx, sub.UserID)
	if err != nil {
		lg.Info("renewOne: no default payment method, recording failure", zap.String("subscriptionId", sub.SubscriptionID.String()))
		s.applyFailure(ctx, sub, "no default payment method on file")
		return
	}

	result, err := s.gateway.Charge(ctx, port.ChargeRequest{
		PaymentMethod: pm,
		Amount:        sub.Amount,
		Currency:      sub.Currency,
		Description:   fmt.Sprintf("Renewal payment for subscription %s", sub.SubscriptionID),
	})
	if err != nil {
		lg.Error("renewOne: gateway unreachable", zap.String("subscriptionId", sub.SubscriptionID.String()), zap.Error(err))
		s.applyFailure(ctx, sub, err.Error())
		return
	}
	if !result.Success {
		s.applyFailure(ctx, sub, result.ErrorMessage)
		return
	}

	if err := s.recordSuccessfulCharge(ctx, sub, pm, result); err != nil {
		lg.Error("renewOne: record charge", zap.Error(err))
	}

	sub.PaymentFailureCount = 0
	sub.NextBillingDate = sub.NextBillingDate.Add(billingPeriod)
	if sub.Status == model.SubscriptionTrial {
		sub.Status = model.SubscriptionActive
		sub.TrialEndDate = nil
	}
	if err := s.subs.Update(ctx, sub); err != nil {
		lg.Error("renewOne: update subscription after successful charge", zap.Error(err))
	}
}

func (s *subscriptionService) applyFailure(ctx context.Context, sub *model.Subscription, reason string) {
	lg := logger.WithContext(ctx)

	failedTx := &model.Transaction{
		TransactionID:  uuid.New(),
		SubscriptionID: sub.SubscriptionID,
		Amount:         sub.Amount,
		Currency:       sub.Currency,
		Status:         model.TransactionFailed,
		ErrorMessage:   reason,
	}
	if err := s.txs.Create(ctx, failedTx); err != nil {
		lg.Error("applyFailure: record failed transaction", zap.Error(err))
	}

	sub.PaymentFailureCount++
	if sub.PaymentFailureCount >= MaxPaymentRetries {
		sub.Status = model.SubscriptionSuspended
		sub.AutoRenewal = false
		lg.Info("applyFailure: subscription suspended after repeated failures", zap.String("subscriptionId", sub.SubscriptionID.String()))
	}
	// NextBillingDate is deliberately left in the past (unless suspended) so
	// the next scheduler run retries the charge.
	if err := s.subs.Update(ctx, sub); err != nil {
		lg.Error("applyFailure: update subscription", zap.Error(err))
	}
}

func (s *subscriptionService) recordSuccessfulCharge(ctx context.Context, sub *model.Subscription, pm *model.PaymentMethod, result port.ChargeResult) error {
	tx := &model.Transaction{
		TransactionID:        uuid.New(),
		SubscriptionID:       sub.SubscriptionID,
		PaymentMethodID:      pm.PaymentMethodID,
		Amount:               sub.Amount,
		Currency:             sub.Currency,
		Status:               model.TransactionSuccessful,
		GatewayTransactionID: result.GatewayTransactionID,
	}
	if err := s.txs.Create(ctx, tx); err != nil {
		return err
	}

	now := time.Now()
	invoice := &model.Invoice{
		InvoiceID:      uuid.New(),
		TransactionID:  tx.TransactionID,
		SubscriptionID: sub.SubscriptionID,
		UserID:         sub.UserID,
		Amount:         sub.Amount,
		TaxAmount:      0,
		TotalAmount:    sub.Amount,
		Status:         model.InvoicePaid,
		DueDate:        now,
		PaidDate:       &now,
		InvoiceNumber:  fmt.Sprintf("INV-%s-%s", now.Format("20060102"), sub.SubscriptionID.String()[:8]),
	}
	return s.inv.Create(ctx, invoice)
}

func (s *subscriptionService) resolvePaymentMethod(ctx context.Context, userID, paymentMethodID uuid.UUID) (*model.PaymentMethod, error) {
	if paymentMethodID != uuid.Nil {
		pm, err := s.pms.GetByID(ctx, paymentMethodID)
		if err != nil {
			return nil, common.ErrPaymentMethodRequired
		}
		return pm, nil
	}
	pm, err := s.pms.GetDefaultByUser(ctx, userID)
	if err != nil {
		return nil, common.ErrPaymentMethodRequired
	}
	return pm, nil
}
