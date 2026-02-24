package usecase

import (
	"context"
	"fmt"

	"github.com/JIeeiroSst/order-processing-service/internal/domain/entity"
	"github.com/JIeeiroSst/order-processing-service/internal/domain/repository"
	"github.com/JIeeiroSst/order-processing-service/internal/proxy"
	apperrors "github.com/JIeeiroSst/order-processing-service/pkg/errors"
)

type orderUseCaseImpl struct {
	orderRepo      repository.OrderRepository
	paymentProxy   proxy.PaymentProxy
	inventoryProxy proxy.InventoryProxy
	shippingProxy  proxy.ShippingProxy
	notifyProxy    proxy.NotificationProxy
}

func NewOrderUseCase(
	orderRepo repository.OrderRepository,
	paymentProxy proxy.PaymentProxy,
	inventoryProxy proxy.InventoryProxy,
	shippingProxy proxy.ShippingProxy,
	notifyProxy proxy.NotificationProxy,
) OrderUseCase {
	return &orderUseCaseImpl{
		orderRepo:      orderRepo,
		paymentProxy:   paymentProxy,
		inventoryProxy: inventoryProxy,
		shippingProxy:  shippingProxy,
		notifyProxy:    notifyProxy,
	}
}

func (uc *orderUseCaseImpl) ValidateOrder(ctx context.Context, order entity.Order) error {
	if order.ID == "" {
		return fmt.Errorf("order ID is required")
	}
	if order.CustomerID == "" {
		return fmt.Errorf("customer ID is required")
	}
	if len(order.Items) == 0 {
		return fmt.Errorf("order must have at least one item")
	}
	if order.TotalAmount <= 0 {
		return fmt.Errorf("order total must be greater than zero")
	}
	return nil
}

func (uc *orderUseCaseImpl) ProcessPayment(ctx context.Context, req entity.PaymentRequest) (*entity.PaymentResponse, error) {
	resp, err := uc.paymentProxy.ChargePayment(ctx, req)
	if err != nil {
		return nil, &apperrors.PaymentFailedError{OrderID: req.OrderID, Reason: err.Error()}
	}
	return resp, nil
}

func (uc *orderUseCaseImpl) ReserveInventory(ctx context.Context, req entity.InventoryReserveRequest) (*entity.InventoryReserveResponse, error) {
	resp, err := uc.inventoryProxy.ReserveStock(ctx, req)
	if err != nil {
		return nil, &apperrors.InventoryError{ProductID: "batch", Reason: err.Error()}
	}
	if !resp.Reserved {
		return nil, &apperrors.InventoryError{ProductID: "batch", Reason: resp.Message}
	}
	return resp, nil
}

func (uc *orderUseCaseImpl) CreateShipment(ctx context.Context, req entity.ShippingRequest) (*entity.ShippingResponse, error) {
	resp, err := uc.shippingProxy.CreateShipment(ctx, req)
	if err != nil {
		return nil, &apperrors.ShippingError{OrderID: req.OrderID, Reason: err.Error()}
	}
	return resp, nil
}

func (uc *orderUseCaseImpl) SendNotification(ctx context.Context, req entity.NotificationRequest) (*entity.NotificationResponse, error) {
	resp, err := uc.notifyProxy.Send(ctx, req)
	if err != nil {
		return &entity.NotificationResponse{Sent: false}, nil
	}
	return resp, nil
}

func (uc *orderUseCaseImpl) UpdateOrderStatus(ctx context.Context, orderID, status string) error {
	return uc.orderRepo.UpdateStatus(ctx, orderID, status)
}

func (uc *orderUseCaseImpl) CancelOrder(ctx context.Context, orderID, reason string) error {
	return uc.orderRepo.UpdateStatus(ctx, orderID, "CANCELLED")
}

func (uc *orderUseCaseImpl) CleanupStaleOrders(ctx context.Context, olderThanMinutes int) (int, error) {
	staleOrders, err := uc.orderRepo.GetStaleOrders(ctx, olderThanMinutes)
	if err != nil {
		return 0, err
	}
	count := 0
	for _, so := range staleOrders {
		if err := uc.orderRepo.UpdateStatus(ctx, so.OrderID, "CANCELLED"); err == nil {
			count++
		}
	}
	return count, nil
}
