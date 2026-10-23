package mysql

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/JIeeiroSst/order-processing-service/internal/domain/entity"
	"github.com/JIeeiroSst/order-processing-service/internal/domain/repository"
	"gorm.io/gorm"
)

type orderRow struct {
	ID                int       `gorm:"column:id;primaryKey"`
	Number            string    `gorm:"column:number;size:64;uniqueIndex"`
	Currency          string    `gorm:"column:currency;size:3"`
	TotalInclTax      float64   `gorm:"column:total_incl_tax"`
	TotalExclTax      float64   `gorm:"column:total_excl_tax"`
	ShippingInclTax   float64   `gorm:"column:shipping_incl_tax"`
	ShippingExclTax   float64   `gorm:"column:shipping_excl_tax"`
	ShippingMethod    string    `gorm:"column:shipping_method"`
	ShippingCode      string    `gorm:"column:shipping_code"`
	CarrierFee        float64   `gorm:"column:shipping_carrier_fee"`
	Status            string    `gorm:"column:status;size:32"`
	GuestEmail        string    `gorm:"column:guest_email"`
	SiteID            int       `gorm:"column:site_id"`
	UserID            int       `gorm:"column:user_id;index"`
	OwnerID           int       `gorm:"column:owner_id"`
	BasketID          int       `gorm:"column:basket_id;uniqueIndex"`
	PaymentReference  string    `gorm:"column:payment_reference;size:64"`
	ShippingAddressID int       `gorm:"column:shipping_address_id"`
	DatePlaced        time.Time `gorm:"column:date_placed"`
}

func (orderRow) TableName() string { return "order_order" }

type orderLineRow struct {
	ID            int     `gorm:"column:id;primaryKey"`
	OrderID       int     `gorm:"column:order_id;index"`
	ProductID     int     `gorm:"column:product_id"`
	StockrecordID int     `gorm:"column:stockrecord_id"`
	Title         string  `gorm:"column:title"`
	Quantity      int     `gorm:"column:quantity"`
	UnitPrice     float64 `gorm:"column:unit_price_incl_tax"`
	LinePrice     float64 `gorm:"column:line_price_incl_tax"`
}

func (orderLineRow) TableName() string { return "order_line" }

type store struct{ db *gorm.DB }

func NewPaidOrderStore(db *gorm.DB) (repository.PaidOrderStore, error) {
	if err := db.AutoMigrate(&orderRow{}, &orderLineRow{}); err != nil {
		return nil, fmt.Errorf("migrate order tables: %w", err)
	}
	return &store{db: db}, nil
}

func (s *store) CreatePaidOrder(ctx context.Context, in entity.PaidOrderInput) (*entity.PaidOrderRecord, error) {
	number := in.Number()
	var existing orderRow
	err := s.db.WithContext(ctx).Where("number = ?", number).First(&existing).Error
	if err == nil {
		return &entity.PaidOrderRecord{ID: existing.ID, Number: existing.Number}, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	row := orderRow{
		Number:            number,
		Currency:          in.Currency,
		TotalInclTax:      in.Total,
		TotalExclTax:      in.Total,
		ShippingInclTax:   in.ShippingFee,
		ShippingExclTax:   in.ShippingFee,
		Status:            "Paid",
		UserID:            in.UserID,
		OwnerID:           in.UserID,
		BasketID:          in.BasketID,
		PaymentReference:  in.PaymentReference,
		ShippingAddressID: in.ShippingAddressID,
		DatePlaced:        time.Now().UTC(),
	}
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&row).Error; err != nil {
			return err
		}
		lines := make([]orderLineRow, len(in.Lines))
		for i, l := range in.Lines {
			lines[i] = orderLineRow{
				OrderID: row.ID, ProductID: l.ProductID, StockrecordID: l.StockrecordID, Title: l.Title,
				Quantity: l.Quantity, UnitPrice: l.UnitPrice, LinePrice: l.UnitPrice * float64(l.Quantity),
			}
		}
		return tx.Create(&lines).Error
	})
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return s.CreatePaidOrder(ctx, in)
		}
		return nil, err
	}
	return &entity.PaidOrderRecord{ID: row.ID, Number: number, Created: true}, nil
}

func (s *store) SetStatus(ctx context.Context, number, status string) error {
	res := s.db.WithContext(ctx).Model(&orderRow{}).Where("number = ?", number).Update("status", status)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("order %s not found", number)
	}
	return nil
}

const (
	StatusAwaitingPickup = "Awaiting pickup"
	StatusNeedsShipping  = "Needs shipping"
)

func (s *store) SetShipment(ctx context.Context, number string, shipment entity.ShipmentResult) error {
	updates := map[string]any{"shipping_method": "GHN"}
	switch shipment.Outcome {
	case entity.ShipmentOutcomeCreated:
		updates["shipping_code"] = shipment.Code
		updates["shipping_carrier_fee"] = float64(shipment.Fee)
		updates["status"] = StatusAwaitingPickup
	default:
		if shipment.Code != "" {
			updates["shipping_code"] = shipment.Code
		}
		updates["status"] = StatusNeedsShipping
	}
	res := s.db.WithContext(ctx).Model(&orderRow{}).Where("number = ?", number).Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("order %s not found", number)
	}
	return nil
}
