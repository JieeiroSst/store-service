package repository

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/shopify-service/internal/domain/model"
	"github.com/JIeeiroSst/shopify-service/internal/domain/port"
	"gorm.io/gorm"
)

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) port.OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) Upsert(ctx context.Context, order *model.Order) (*model.Order, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing model.Order
		err := tx.Where("shopify_id = ?", order.ShopifyID).First(&existing).Error
		switch {
		case err == nil:
			order.ID = existing.ID
			order.CreatedAt = existing.CreatedAt
		case !errors.Is(err, gorm.ErrRecordNotFound):
			return err
		}

		items := order.LineItems
		order.LineItems = nil
		if err := tx.Omit("LineItems").Save(order).Error; err != nil {
			return err
		}

		if err := tx.Where("order_id = ?", order.ID).Delete(&model.LineItem{}).Error; err != nil {
			return err
		}
		for i := range items {
			items[i].ID = 0
			items[i].OrderID = order.ID
		}
		if len(items) > 0 {
			if err := tx.Create(&items).Error; err != nil {
				return err
			}
		}
		order.LineItems = items
		return nil
	})
	if err != nil {
		return nil, err
	}
	return order, nil
}

func (r *orderRepository) GetByID(ctx context.Context, id int64) (*model.Order, error) {
	var order model.Order
	if err := r.db.WithContext(ctx).Preload("LineItems").First(&order, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, port.ErrNotFound
		}
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) List(ctx context.Context, limit, offset int) ([]model.Order, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.Order{})

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var orders []model.Order
	if err := query.Preload("LineItems").Order("processed_at DESC").Limit(limit).Offset(offset).Find(&orders).Error; err != nil {
		return nil, 0, err
	}
	return orders, total, nil
}
