package internalgateway

import (
	"context"
	"time"

	paymentapp "github.com/JIeeiroSst/voucher-service/internal/application/payment"
	reconciliationapp "github.com/JIeeiroSst/voucher-service/internal/application/reconciliation"
)

type PaymentRecordSourceAdapter struct {
	paymentSvc paymentapp.PaymentService
}

func NewPaymentRecordSourceAdapter(paymentSvc paymentapp.PaymentService) reconciliationapp.PaymentRecordSource {
	return &PaymentRecordSourceAdapter{paymentSvc: paymentSvc}
}

func (a *PaymentRecordSourceAdapter) ListSettledSince(ctx context.Context, since time.Time) ([]reconciliationapp.PaymentRecord, error) {
	payments, err := a.paymentSvc.ListSettledSince(ctx, since)
	if err != nil {
		return nil, err
	}
	records := make([]reconciliationapp.PaymentRecord, 0, len(payments))
	for _, p := range payments {
		records = append(records, reconciliationapp.PaymentRecord{
			PaymentID:      p.ID,
			OrderID:        p.OrderID.String(),
			Amount:         p.Amount.Amount,
			Currency:       p.Amount.Currency,
			ProviderTxnRef: p.ProviderTxnRef,
		})
	}
	return records, nil
}
