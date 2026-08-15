package internalgateway

import (
	"context"

	orderapp "github.com/JIeeiroSst/voucher-service/internal/application/order"
	paymentapp "github.com/JIeeiroSst/voucher-service/internal/application/payment"
)

type PaymentInitiatorAdapter struct {
	paymentSvc paymentapp.PaymentService
}

func NewPaymentInitiatorAdapter(paymentSvc paymentapp.PaymentService) orderapp.PaymentInitiator {
	return &PaymentInitiatorAdapter{paymentSvc: paymentSvc}
}

func (a *PaymentInitiatorAdapter) InitiatePayment(ctx context.Context, req orderapp.PaymentRequest) (orderapp.PaymentIntent, error) {
	out, err := a.paymentSvc.InitiatePayment(ctx, paymentapp.InitiatePaymentInput{
		OrderID:   req.OrderID,
		Amount:    req.Amount,
		Provider:  req.Method,
		ReturnURL: req.ReturnURL,
	})
	if err != nil {
		return orderapp.PaymentIntent{}, err
	}
	return orderapp.PaymentIntent{PaymentID: out.PaymentID, RedirectURL: out.RedirectURL}, nil
}
