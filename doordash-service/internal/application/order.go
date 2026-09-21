package application

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/JIeeiroSst/doordash-service/internal/domain/model"
	"github.com/JIeeiroSst/doordash-service/internal/domain/port"
)

type orderService struct {
	orders      port.OrderRepository
	tracking    port.OrderTrackingRepository
	users       port.UserClient
	restaurants port.RestaurantClient
	payments    port.PaymentClient
	notifier    port.NotifierClient
}

func NewOrderService(
	orders port.OrderRepository,
	tracking port.OrderTrackingRepository,
	users port.UserClient,
	restaurants port.RestaurantClient,
	payments port.PaymentClient,
	notifier port.NotifierClient,
) port.OrderUsecase {
	return &orderService{
		orders: orders, tracking: tracking,
		users: users, restaurants: restaurants, payments: payments, notifier: notifier,
	}
}

func (s *orderService) CreateOrder(ctx context.Context, in port.CreateOrderInput) (*model.Order, error) {
	if len(in.Items) == 0 {
		return nil, port.ErrEmptyOrder
	}

	restaurant, err := s.restaurants.GetRestaurant(ctx, in.RestaurantID)
	if err != nil {
		return nil, fmt.Errorf("fetch restaurant: %w", err)
	}
	if !restaurant.IsActive {
		return nil, port.ErrRestaurantInactive
	}

	menuItemIDs := make([]string, len(in.Items))
	for i, item := range in.Items {
		menuItemIDs[i] = item.MenuItemID
	}
	menuItems, err := s.restaurants.GetMenuItems(ctx, menuItemIDs)
	if err != nil {
		return nil, fmt.Errorf("fetch menu items: %w", err)
	}

	items := make([]model.OrderItem, len(in.Items))
	for i, in := range in.Items {
		menuItem, ok := menuItems[in.MenuItemID]
		if !ok || !menuItem.IsActive {
			return nil, port.ErrMenuItemUnavailable
		}
		if in.Quantity <= 0 {
			return nil, fmt.Errorf("%w: item %s has non-positive quantity", port.ErrMenuItemUnavailable, in.MenuItemID)
		}
		items[i] = model.OrderItem{
			MenuItemID:          menuItem.ID,
			Name:                menuItem.Name,
			Quantity:            in.Quantity,
			UnitPrice:           menuItem.Price,
			Customizations:      in.Customizations,
			SpecialInstructions: in.SpecialInstructions,
		}
	}

	if _, err := s.users.GetCustomer(ctx, in.CustomerID); err != nil {
		return nil, fmt.Errorf("fetch customer: %w", err)
	}

	order := &model.Order{
		CustomerID:          in.CustomerID,
		RestaurantID:        in.RestaurantID,
		DeliveryAddressID:   in.DeliveryAddressID,
		Status:              model.OrderStatusCreated,
		PlacedAt:            time.Now(),
		DeliveryFee:         in.DeliveryFee,
		ServiceFee:          in.ServiceFee,
		Tax:                 in.Tax,
		Tip:                 in.Tip,
		PaymentMethodID:     in.PaymentMethodID,
		PaymentStatus:       model.PaymentStatusPending,
		SpecialInstructions: in.SpecialInstructions,
		Items:               items,
	}
	order.Recalculate()

	authorization, err := s.payments.AuthorizePayment(ctx, port.AuthorizePaymentInput{
		CustomerID:      in.CustomerID,
		Amount:          order.TotalAmount,
		PaymentMethodID: in.PaymentMethodID,
	})
	if err != nil || authorization.Status != model.PaymentStatusAuthorized {
		return nil, port.ErrPaymentAuthorizationFailed
	}
	order.PaymentStatus = model.PaymentStatusAuthorized

	created, err := s.orders.Create(ctx, order)
	if err != nil {
		return nil, fmt.Errorf("persist order: %w", err)
	}

	if err := s.tracking.Create(ctx, &model.OrderTracking{
		OrderID:   created.ID,
		Status:    model.OrderStatusCreated,
		Timestamp: created.PlacedAt,
	}); err != nil {
		log.Printf("doordash: record initial tracking event for order %d: %v", created.ID, err)
	}

	s.notifyBestEffort(ctx, port.NotifyInput{
		UserID:      created.CustomerID,
		Title:       "Order placed",
		Message:     fmt.Sprintf("Your order #%d has been placed.", created.ID),
		Type:        "order_update",
		ReferenceID: fmt.Sprint(created.ID),
	})

	return created, nil
}

func (s *orderService) GetOrder(ctx context.Context, id int64) (*model.Order, error) {
	return s.orders.GetByID(ctx, id)
}

func (s *orderService) ListOrdersByCustomer(ctx context.Context, customerID string) ([]model.Order, error) {
	return s.orders.ListByCustomer(ctx, customerID)
}

func (s *orderService) ListOrdersByRestaurant(ctx context.Context, restaurantID string) ([]model.Order, error) {
	return s.orders.ListByRestaurant(ctx, restaurantID)
}

func (s *orderService) UpdateOrderStatus(ctx context.Context, id int64, status model.OrderStatus) (*model.Order, error) {
	if !status.Valid() {
		return nil, fmt.Errorf("%w: %q", port.ErrInvalidStatusTransition, status)
	}

	order, err := s.orders.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !order.Status.CanTransitionTo(status) {
		return nil, fmt.Errorf("%w: %s -> %s", port.ErrInvalidStatusTransition, order.Status, status)
	}

	if err := s.orders.UpdateStatus(ctx, id, status); err != nil {
		return nil, err
	}
	if status == model.OrderStatusDelivered {
		if err := s.orders.SetActualDeliveryTime(ctx, id, time.Now()); err != nil {
			log.Printf("doordash: set actual delivery time for order %d: %v", id, err)
		}
	}

	if err := s.tracking.Create(ctx, &model.OrderTracking{
		OrderID:   id,
		Status:    status,
		Timestamp: time.Now(),
	}); err != nil {
		log.Printf("doordash: record tracking event for order %d: %v", id, err)
	}

	s.notifyBestEffort(ctx, port.NotifyInput{
		UserID:      order.CustomerID,
		Title:       "Order update",
		Message:     fmt.Sprintf("Your order #%d is now %s.", id, status),
		Type:        "order_update",
		ReferenceID: fmt.Sprint(id),
	})

	return s.orders.GetByID(ctx, id)
}

func (s *orderService) CancelOrder(ctx context.Context, id int64, reason string) (*model.Order, error) {
	order, err := s.orders.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if order.Status.Terminal() {
		return nil, port.ErrOrderAlreadyTerminal
	}
	if !order.Status.CanTransitionTo(model.OrderStatusCancelled) {
		return nil, fmt.Errorf("%w: %s -> %s", port.ErrInvalidStatusTransition, order.Status, model.OrderStatusCancelled)
	}

	if err := s.orders.UpdateStatus(ctx, id, model.OrderStatusCancelled); err != nil {
		return nil, err
	}
	if err := s.tracking.Create(ctx, &model.OrderTracking{
		OrderID:   id,
		Status:    model.OrderStatusCancelled,
		Timestamp: time.Now(),
		Notes:     reason,
	}); err != nil {
		log.Printf("doordash: record cancellation tracking event for order %d: %v", id, err)
	}

	s.notifyBestEffort(ctx, port.NotifyInput{
		UserID:      order.CustomerID,
		Title:       "Order cancelled",
		Message:     fmt.Sprintf("Your order #%d was cancelled: %s", id, reason),
		Type:        "order_update",
		ReferenceID: fmt.Sprint(id),
	})

	return s.orders.GetByID(ctx, id)
}

// notifyBestEffort sends a customer notification without letting a
// notification-service outage fail the order operation that triggered it.
func (s *orderService) notifyBestEffort(ctx context.Context, in port.NotifyInput) {
	if err := s.notifier.Notify(ctx, in); err != nil {
		log.Printf("doordash: notify user %s: %v", in.UserID, err)
	}
}
