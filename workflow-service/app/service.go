package app

import (
	"errors"
	"fmt"
)

type OrderService struct {
	repo          OrderRepository
	paymentClient PaymentClient
	fulfillment   FulfillmentService
	notifier      NotificationService
}

func NewOrderService(
	repo OrderRepository,
	paymentClient PaymentClient,
	fulfillment FulfillmentService,
	notifier NotificationService,
) *OrderService {
	return &OrderService{
		repo:          repo,
		paymentClient: paymentClient,
		fulfillment:   fulfillment,
		notifier:      notifier,
	}
}

func (s *OrderService) ProcessOrder(orderID string) error {
	order, err := s.repo.GetOrder(orderID)
	if err != nil {
		return fmt.Errorf("failed to retrieve order: %w", err)
	}

	if order.Status == "pending" {
		return errors.New("order is not in pending status")
	}

	paymentID, err := s.paymentClient.ProcessPayment(orderID, order.Amount)
	if err != nil {
		return fmt.Errorf("payment processing failed: %w", err)
	}

	order.PaymentID = paymentID
	order.Status = "paid"
	if err := s.repo.UpdateOrder(order); err != nil {
		_ = s.paymentClient.RefundPayment(paymentID) 
		return fmt.Errorf("failed to update order after payment: %w", err)
	}

	trackingNumber, err := s.fulfillment.FulfillOrder(orderID)
	if err != nil {
		_ = s.paymentClient.RefundPayment(paymentID)
		order.Status = "fulfillment_failed"
		_ = s.repo.UpdateOrder(order)

		return fmt.Errorf("order fulfillment failed: %w", err)
	}

	order.TrackingNumber = trackingNumber
	order.Status = "fulfilled"
	if err := s.repo.UpdateOrder(order); err != nil {
		return fmt.Errorf("failed to update order after fulfillment: %w", err)
	}

	err = s.notifier.SendOrderConfirmation(orderID, trackingNumber)
	if err != nil {
		fmt.Printf("Failed to send confirmation for order %s: %v\n", orderID, err)
	}

	return nil
}
