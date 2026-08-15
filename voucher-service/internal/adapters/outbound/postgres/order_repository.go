package postgres

import (
	"context"
	"encoding/json"
	"time"

	orderapp "github.com/JIeeiroSst/voucher-service/internal/application/order"
	"github.com/JIeeiroSst/voucher-service/internal/domain/order"
	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
	"github.com/JIeeiroSst/voucher-service/internal/platform/txmanager"
	"gorm.io/gorm"
)

type orderModel struct {
	ID             string    `gorm:"column:id;primaryKey"`
	BuyerType      string    `gorm:"column:buyer_type"`
	BuyerID        string    `gorm:"column:buyer_id"`
	Status         string    `gorm:"column:status"`
	TotalAmount    float64   `gorm:"column:total_amount"`
	Currency       string    `gorm:"column:currency"`
	Version        int       `gorm:"column:version"`
	IdempotencyKey *string   `gorm:"column:idempotency_key"`
	PaymentRef     string    `gorm:"column:payment_ref"`
	CreatedAt      time.Time `gorm:"column:created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at"`
}

func (orderModel) TableName() string { return "orders" }

type orderItemModel struct {
	ID               string  `gorm:"column:id;primaryKey"`
	OrderID          string  `gorm:"column:order_id"`
	MerchantID       string  `gorm:"column:merchant_id"`
	ProductSKU       string  `gorm:"column:product_sku"`
	Quantity         int     `gorm:"column:quantity"`
	UnitPrice        float64 `gorm:"column:unit_price"`
	LineTotal        float64 `gorm:"column:line_total"`
	IssuedVoucherIDs []byte  `gorm:"column:issued_voucher_ids"`
}

func (orderItemModel) TableName() string { return "order_items" }

func orderToModels(o *order.Order) (*orderModel, []orderItemModel, error) {
	om := &orderModel{
		ID:          o.ID.String(),
		BuyerType:   string(o.BuyerType),
		BuyerID:     o.BuyerID,
		Status:      string(o.Status),
		TotalAmount: float64(o.TotalAmount.Amount),
		Currency:    o.TotalAmount.Currency,
		Version:     o.Version,
		PaymentRef:  o.PaymentRef,
		CreatedAt:   o.CreatedAt,
		UpdatedAt:   o.UpdatedAt,
	}
	if o.IdempotencyKey != "" {
		om.IdempotencyKey = &o.IdempotencyKey
	}

	items := make([]orderItemModel, 0, len(o.Items))
	for _, item := range o.Items {
		voucherIDs := make([]string, 0, len(item.IssuedVoucherIDs))
		for _, id := range item.IssuedVoucherIDs {
			voucherIDs = append(voucherIDs, id.String())
		}
		idsJSON, err := json.Marshal(voucherIDs)
		if err != nil {
			return nil, nil, err
		}
		items = append(items, orderItemModel{
			ID:               item.ID.String(),
			OrderID:          o.ID.String(),
			MerchantID:       item.MerchantID.String(),
			ProductSKU:       item.ProductSKU,
			Quantity:         item.Quantity,
			UnitPrice:        float64(item.UnitPrice.Amount),
			LineTotal:        float64(item.LineTotal.Amount),
			IssuedVoucherIDs: idsJSON,
		})
	}
	return om, items, nil
}

func orderFromModels(om *orderModel, items []orderItemModel) (*order.Order, error) {
	id, err := shared.ParseOrderID(om.ID)
	if err != nil {
		return nil, err
	}
	o := &order.Order{
		ID:               id,
		BuyerType:        order.BuyerType(om.BuyerType),
		BuyerID:          om.BuyerID,
		Status:           order.Status(om.Status),
		TotalAmount:      shared.NewMoney(int64(om.TotalAmount), om.Currency),
		Version:          om.Version,
		PersistedVersion: om.Version,
		PaymentRef:       om.PaymentRef,
		CreatedAt:        om.CreatedAt,
		UpdatedAt:        om.UpdatedAt,
	}
	if om.IdempotencyKey != nil {
		o.IdempotencyKey = *om.IdempotencyKey
	}

	for _, im := range items {
		itemID, err := shared.ParseOrderID(im.ID)
		if err != nil {
			return nil, err
		}
		merchantID, err := shared.ParseMerchantID(im.MerchantID)
		if err != nil {
			return nil, err
		}
		var voucherIDStrs []string
		if len(im.IssuedVoucherIDs) > 0 {
			if err := json.Unmarshal(im.IssuedVoucherIDs, &voucherIDStrs); err != nil {
				return nil, err
			}
		}
		voucherIDs := make([]shared.VoucherID, 0, len(voucherIDStrs))
		for _, s := range voucherIDStrs {
			vid, err := shared.ParseVoucherID(s)
			if err != nil {
				return nil, err
			}
			voucherIDs = append(voucherIDs, vid)
		}
		o.Items = append(o.Items, order.OrderItem{
			ID:               itemID,
			MerchantID:       merchantID,
			ProductSKU:       im.ProductSKU,
			Quantity:         im.Quantity,
			UnitPrice:        shared.NewMoney(int64(im.UnitPrice), om.Currency),
			LineTotal:        shared.NewMoney(int64(im.LineTotal), om.Currency),
			IssuedVoucherIDs: voucherIDs,
		})
	}
	return o, nil
}

type OrderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) orderapp.OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Create(ctx context.Context, o *order.Order) error {
	om, items, err := orderToModels(o)
	if err != nil {
		return err
	}
	return txmanager.DBFromContext(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(om).Error; err != nil {
			return err
		}
		if len(items) > 0 {
			if err := tx.Create(&items).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *OrderRepository) find(ctx context.Context, id shared.OrderID, forUpdate bool) (*order.Order, error) {
	db := txmanager.DBFromContext(ctx, r.db)
	var om orderModel
	q := db
	if forUpdate {
		q = lockForUpdate(db)
	}
	if err := q.First(&om, "id = ?", id.String()).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, order.ErrOrderNotFound
		}
		return nil, err
	}
	var items []orderItemModel
	if err := db.Where("order_id = ?", id.String()).Find(&items).Error; err != nil {
		return nil, err
	}
	return orderFromModels(&om, items)
}

func (r *OrderRepository) FindByID(ctx context.Context, id shared.OrderID) (*order.Order, error) {
	return r.find(ctx, id, false)
}

func (r *OrderRepository) FindByIDForUpdate(ctx context.Context, id shared.OrderID) (*order.Order, error) {
	return r.find(ctx, id, true)
}

func (r *OrderRepository) ListByBuyer(ctx context.Context, buyerID string) ([]*order.Order, error) {
	var oms []orderModel
	if err := txmanager.DBFromContext(ctx, r.db).Where("buyer_id = ?", buyerID).Find(&oms).Error; err != nil {
		return nil, err
	}
	out := make([]*order.Order, 0, len(oms))
	for i := range oms {
		var items []orderItemModel
		if err := r.db.WithContext(ctx).Where("order_id = ?", oms[i].ID).Find(&items).Error; err != nil {
			return nil, err
		}
		o, err := orderFromModels(&oms[i], items)
		if err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, nil
}

func (r *OrderRepository) Save(ctx context.Context, o *order.Order) error {
	om, items, err := orderToModels(o)
	if err != nil {
		return err
	}
	return txmanager.DBFromContext(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&orderModel{}).
			Where("id = ? AND version = ?", om.ID, o.PersistedVersion).
			Updates(map[string]any{
				"status":       om.Status,
				"total_amount": om.TotalAmount,
				"version":      om.Version,
				"payment_ref":  om.PaymentRef,
				"updated_at":   om.UpdatedAt,
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return order.ErrVersionConflict
		}
		for _, item := range items {
			if err := tx.Model(&orderItemModel{}).
				Where("id = ?", item.ID).
				Update("issued_voucher_ids", item.IssuedVoucherIDs).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
