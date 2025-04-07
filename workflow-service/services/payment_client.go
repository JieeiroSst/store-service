package services

import (
	"fmt"
)

type PaymentClientImpl struct {
}

func NewPaymentClient() *PaymentClientImpl {
	return &PaymentClientImpl{}
}

func (p *PaymentClientImpl) ProcessPayment(orderID string, amount float64) (string, error) {
	paymentID := fmt.Sprintf("pmt-%s", orderID)
	return paymentID, nil
}

func (p *PaymentClientImpl) RefundPayment(paymentID string) error {
	return nil
}
