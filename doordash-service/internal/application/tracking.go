package application

import (
	"context"
	"time"

	"github.com/JIeeiroSst/doordash-service/internal/domain/model"
	"github.com/JIeeiroSst/doordash-service/internal/domain/port"
)

type trackingService struct {
	orders   port.OrderRepository
	tracking port.OrderTrackingRepository
}

func NewTrackingService(orders port.OrderRepository, tracking port.OrderTrackingRepository) port.TrackingUsecase {
	return &trackingService{orders: orders, tracking: tracking}
}

// AddTrackingEvent records a manual checkpoint (e.g. a driver's live GPS
// ping) against an existing order. Status-changing checkpoints go through
// OrderUsecase.UpdateOrderStatus instead, which also validates the
// transition and updates the order row itself.
func (s *trackingService) AddTrackingEvent(ctx context.Context, in port.AddTrackingEventInput) (*model.OrderTracking, error) {
	if _, err := s.orders.GetByID(ctx, in.OrderID); err != nil {
		return nil, err
	}
	if !in.Status.Valid() {
		return nil, port.ErrInvalidStatusTransition
	}

	event := &model.OrderTracking{
		OrderID:   in.OrderID,
		Status:    in.Status,
		Latitude:  in.Latitude,
		Longitude: in.Longitude,
		Timestamp: time.Now(),
		Notes:     in.Notes,
	}
	if err := s.tracking.Create(ctx, event); err != nil {
		return nil, err
	}
	return event, nil
}

func (s *trackingService) ListTrackingByOrder(ctx context.Context, orderID int64) ([]model.OrderTracking, error) {
	return s.tracking.ListByOrder(ctx, orderID)
}
