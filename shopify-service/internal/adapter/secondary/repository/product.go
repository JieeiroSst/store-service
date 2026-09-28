package repository

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/shopify-service/internal/domain/model"
	"github.com/JIeeiroSst/shopify-service/internal/domain/port"
	"gorm.io/gorm"
)

type productRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) port.ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) Upsert(ctx context.Context, product *model.Product) (*model.Product, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing model.Product
		err := tx.Where("shopify_id = ?", product.ShopifyID).First(&existing).Error
		switch {
		case err == nil:
			product.ID = existing.ID
			product.CreatedAt = existing.CreatedAt
		case !errors.Is(err, gorm.ErrRecordNotFound):
			return err
		}

		variants := product.Variants
		product.Variants = nil
		if err := tx.Omit("Variants").Save(product).Error; err != nil {
			return err
		}

		if err := tx.Where("product_id = ?", product.ID).Delete(&model.ProductVariant{}).Error; err != nil {
			return err
		}
		for i := range variants {
			variants[i].ID = 0
			variants[i].ProductID = product.ID
		}
		if len(variants) > 0 {
			if err := tx.Create(&variants).Error; err != nil {
				return err
			}
		}
		product.Variants = variants
		return nil
	})
	if err != nil {
		return nil, err
	}
	return product, nil
}

func (r *productRepository) GetByID(ctx context.Context, id int64) (*model.Product, error) {
	var product model.Product
	if err := r.db.WithContext(ctx).Preload("Variants").First(&product, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, port.ErrNotFound
		}
		return nil, err
	}
	return &product, nil
}

func (r *productRepository) List(ctx context.Context, limit, offset int) ([]model.Product, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.Product{})

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var products []model.Product
	if err := query.Preload("Variants").Order("id DESC").Limit(limit).Offset(offset).Find(&products).Error; err != nil {
		return nil, 0, err
	}
	return products, total, nil
}

func (r *productRepository) DeleteByShopifyID(ctx context.Context, shopifyID string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var product model.Product
		err := tx.Where("shopify_id = ?", shopifyID).First(&product).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		if err != nil {
			return err
		}
		if err := tx.Where("product_id = ?", product.ID).Delete(&model.ProductVariant{}).Error; err != nil {
			return err
		}
		return tx.Delete(&product).Error
	})
}
