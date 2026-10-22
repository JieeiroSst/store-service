package repository

import (
	"context"
	"time"

	"github.com/JIeeiroSst/catalogues-service/internal/domain/model"
	"github.com/JIeeiroSst/catalogues-service/internal/domain/port"
	"gorm.io/gorm"
)

type stockRecordEntity struct {
	ID            int64   `gorm:"primaryKey"`
	ProductID     int64   `gorm:"not null;index"`
	PartnerSKU    string  `gorm:"column:partner_sku"`
	PriceCurrency string  `gorm:"not null;size:3"`
	Price         float64 `gorm:"not null"`
	NumInStock    *int
	NumAllocated  int       `gorm:"not null;default:0"`
	DateCreated   time.Time `gorm:"autoCreateTime"`
	DateUpdated   time.Time `gorm:"autoUpdateTime"`

	Product productEntity `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE"`
}

func (stockRecordEntity) TableName() string { return "partner_stockrecord" }

func (e stockRecordEntity) toModel() model.StockRecord {
	return model.StockRecord{
		ID: e.ID, ProductID: e.ProductID, PartnerSKU: e.PartnerSKU, PriceCurrency: e.PriceCurrency,
		Price: e.Price, NumInStock: e.NumInStock, NumAllocated: e.NumAllocated,
		DateCreated: e.DateCreated, DateUpdated: e.DateUpdated,
	}
}

type stockRecordRepo struct{ db *gorm.DB }

func NewStockRecordRepo(db *gorm.DB) port.StockRecordRepository { return &stockRecordRepo{db: db} }

func (r *stockRecordRepo) Create(ctx context.Context, s *model.StockRecord) error {
	e := stockRecordEntity{
		ProductID: s.ProductID, PartnerSKU: s.PartnerSKU, PriceCurrency: s.PriceCurrency,
		Price: s.Price, NumInStock: s.NumInStock,
	}
	if err := r.db.WithContext(ctx).Omit("Product").Create(&e).Error; err != nil {
		return writeErr(err)
	}
	s.ID = e.ID
	return nil
}

func (r *stockRecordRepo) Get(ctx context.Context, id int64) (*model.StockRecord, error) {
	var e stockRecordEntity
	if err := r.db.WithContext(ctx).First(&e, id).Error; err != nil {
		return nil, translate(err, err)
	}
	m := e.toModel()
	return &m, nil
}

func (r *stockRecordRepo) ListByProducts(ctx context.Context, productIDs []int64) ([]model.StockRecord, error) {
	var rows []stockRecordEntity
	if err := r.db.WithContext(ctx).Where("product_id IN ?", productIDs).Order("price ASC, id ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]model.StockRecord, len(rows))
	for i, e := range rows {
		out[i] = e.toModel()
	}
	return out, nil
}

func (r *stockRecordRepo) Update(ctx context.Context, s *model.StockRecord) error {
	res := r.db.WithContext(ctx).Model(&stockRecordEntity{ID: s.ID}).Updates(map[string]any{
		"partner_sku": s.PartnerSKU, "price_currency": s.PriceCurrency, "price": s.Price, "num_in_stock": s.NumInStock,
	})
	if res.Error != nil {
		return writeErr(res.Error)
	}
	if res.RowsAffected == 0 {
		return model.ErrNotFound
	}
	return nil
}

func (r *stockRecordRepo) Delete(ctx context.Context, id int64) error {
	return deleteByID[stockRecordEntity](r.db.WithContext(ctx), id)
}

func (r *stockRecordRepo) Allocate(ctx context.Context, id int64, qty int) error {
	res := r.db.WithContext(ctx).Model(&stockRecordEntity{}).
		Where("id = ? AND (num_in_stock IS NULL OR num_in_stock - num_allocated >= ?)", id, qty).
		Update("num_allocated", gorm.Expr("num_allocated + ?", qty))
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		if _, err := r.Get(ctx, id); err != nil {
			return err
		}
		return model.ErrOutOfStock
	}
	return nil
}

func (r *stockRecordRepo) Release(ctx context.Context, id int64, qty int) error {
	res := r.db.WithContext(ctx).Model(&stockRecordEntity{}).Where("id = ?", id).
		Update("num_allocated", gorm.Expr("GREATEST(num_allocated - ?, 0)", qty))
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return model.ErrNotFound
	}
	return nil
}
