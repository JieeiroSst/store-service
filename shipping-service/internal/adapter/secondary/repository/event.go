package repository

import (
	"context"

	"github.com/JIeeiroSst/shipping-service/internal/domain/model"
	"github.com/JIeeiroSst/shipping-service/internal/domain/port"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type eventRepository struct {
	db *gorm.DB
}

func NewEventRepository(db *gorm.DB) port.EventRepository {
	return &eventRepository{db: db}
}

func (r *eventRepository) Create(ctx context.Context, e *model.ShipmentEvent) (bool, error) {
	res := conn(ctx, r.db).Clauses(clause.OnConflict{DoNothing: true}).Create(e)
	return res.RowsAffected == 1, res.Error
}

func (r *eventRepository) ListByShipment(ctx context.Context, shipmentID int64) ([]model.ShipmentEvent, error) {
	var events []model.ShipmentEvent
	err := conn(ctx, r.db).Where("shipment_id = ?", shipmentID).Order("occurred_at ASC, id ASC").Find(&events).Error
	return events, err
}
