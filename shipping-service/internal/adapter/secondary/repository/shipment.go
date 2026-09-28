package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/JIeeiroSst/shipping-service/internal/domain/model"
	"github.com/JIeeiroSst/shipping-service/internal/domain/port"
	"gorm.io/gorm"
)

type shipmentRepository struct {
	db *gorm.DB
}

func NewShipmentRepository(db *gorm.DB) port.ShipmentRepository {
	return &shipmentRepository{db: db}
}

func (r *shipmentRepository) Create(ctx context.Context, s *model.Shipment) error {
	return translate(conn(ctx, r.db).Omit("PlacingUntil").Create(s).Error, "shipment "+s.ClientOrderCode)
}

func (r *shipmentRepository) Save(ctx context.Context, s *model.Shipment) error {
	return conn(ctx, r.db).Omit("PlacingUntil", "CreatedAt").Save(s).Error
}

func (r *shipmentRepository) first(ctx context.Context, what string, lock bool, query string, args ...any) (*model.Shipment, error) {
	var s model.Shipment
	db := conn(ctx, r.db)
	if lock {
		db = db.Clauses(forUpdate)
	}
	if err := db.Where(query, args...).First(&s).Error; err != nil {
		return nil, translate(err, what)
	}
	return &s, nil
}

func (r *shipmentRepository) GetByID(ctx context.Context, id int64) (*model.Shipment, error) {
	return r.first(ctx, fmt.Sprintf("shipment %d", id), false, "id = ?", id)
}

func (r *shipmentRepository) GetByIDForUpdate(ctx context.Context, id int64) (*model.Shipment, error) {
	return r.first(ctx, fmt.Sprintf("shipment %d", id), true, "id = ?", id)
}

func (r *shipmentRepository) GetByClientOrderCode(ctx context.Context, client, clientOrderCode string) (*model.Shipment, error) {
	return r.first(ctx, "shipment "+clientOrderCode, false, "client_service = ? AND client_order_code = ?", client, clientOrderCode)
}

func (r *shipmentRepository) GetByCode(ctx context.Context, code string) (*model.Shipment, error) {
	return r.first(ctx, "shipment "+code, false, "code = ?", code)
}

func (r *shipmentRepository) GetByCarrierOrderCode(ctx context.Context, carrierOrderCode string) (*model.Shipment, error) {
	return r.first(ctx, "shipment "+carrierOrderCode, false, "carrier_order_code = ?", carrierOrderCode)
}

func (r *shipmentRepository) List(ctx context.Context, client string, filter port.ShipmentFilter) ([]model.Shipment, int64, error) {
	q := conn(ctx, r.db).Model(&model.Shipment{}).Where("client_service = ?", client)
	if filter.Status != "" {
		q = q.Where("status = ?", filter.Status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []model.Shipment
	if err := q.Order("id DESC").Limit(filter.Limit).Offset(filter.Offset).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *shipmentRepository) AcquirePlacement(ctx context.Context, id int64, now, until time.Time) (bool, error) {
	res := conn(ctx, r.db).Model(&model.Shipment{}).
		Where("id = ? AND (placing_until IS NULL OR placing_until < ?)", id, now).
		UpdateColumn("placing_until", until)
	return res.RowsAffected == 1, res.Error
}

func (r *shipmentRepository) ReleasePlacement(ctx context.Context, id int64) error {
	return conn(ctx, r.db).Model(&model.Shipment{}).Where("id = ?", id).UpdateColumn("placing_until", nil).Error
}
