package application

import (
	"context"
	"math"

	"github.com/JIeeiroSst/delivery-service/internal/domain/model"
	"github.com/JIeeiroSst/delivery-service/internal/domain/port"
	"github.com/JieeiroSst/logger"
)

// Fee/ETA estimation constants: a flat base fee plus a per-km rate, and an
// ETA derived from an assumed average delivery speed.
const (
	baseFee           = 1.0
	perKmFee          = 0.5
	avgSpeedKmPerHour = 30.0
)

// estimateFeeAndETA is the single source of truth for Delivery.Fee/ETAMinutes
// — always derived from DistanceKm server-side, never trusted from a caller.
func estimateFeeAndETA(distanceKm float64) (fee float64, etaMinutes int) {
	if distanceKm <= 0 {
		return 0, 0
	}
	fee = baseFee + perKmFee*distanceKm
	etaMinutes = int(math.Ceil(distanceKm / avgSpeedKmPerHour * 60))
	return fee, etaMinutes
}

type deliveryService struct {
	repo port.DeliveryRepository
}

func NewDeliveryService(repo port.DeliveryRepository) port.DeliveryUsecase {
	return &deliveryService{repo: repo}
}

func (s *deliveryService) Create(ctx context.Context, delivery *model.Delivery) error {
	delivery.ShipID = logger.GearedIntID()
	delivery.Status = model.StatusFree
	delivery.Fee, delivery.ETAMinutes = estimateFeeAndETA(delivery.DistanceKm)

	return s.repo.Create(ctx, *delivery)
}

func (s *deliveryService) UpdateStatus(ctx context.Context, shipID int, status string) error {
	st := model.StatusFree
	if status == "busy" {
		st = model.StatusBusy
	}

	return s.repo.UpdateStatus(ctx, shipID, st)
}

func (s *deliveryService) FindByActive(ctx context.Context) (*model.Delivery, error) {
	return s.repo.FindByActive(ctx)
}

func (s *deliveryService) FindAll(ctx context.Context, pagination logger.Pagination) (logger.Pagination, error) {
	return s.repo.FindAll(ctx, pagination)
}

func (s *deliveryService) Update(ctx context.Context, shipID int, delivery *model.Delivery) error {
	delivery.Fee, delivery.ETAMinutes = estimateFeeAndETA(delivery.DistanceKm)

	return s.repo.Update(ctx, shipID, *delivery)
}
