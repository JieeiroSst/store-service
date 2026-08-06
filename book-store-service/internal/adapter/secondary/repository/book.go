package repository

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/bookStore-service/common"
	"github.com/JIeeiroSst/bookStore-service/internal/domain/model"
	"github.com/JIeeiroSst/bookStore-service/internal/domain/port"
	"gorm.io/gorm"
)

type bookRepository struct {
	db *gorm.DB
}

func NewBookRepository(db *gorm.DB) port.BookRepository {
	return &bookRepository{db: db}
}

func (r *bookRepository) Create(ctx context.Context, book *model.Book) error {
	if err := r.db.WithContext(ctx).Create(book).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *bookRepository) preloaded(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).Preload("Author").Preload("Publisher").Preload("Category")
}

func (r *bookRepository) GetByID(ctx context.Context, id int) (*model.Book, error) {
	var book model.Book
	if err := r.preloaded(ctx).First(&book, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound
		}
		return nil, common.ErrDBFailed
	}
	return &book, nil
}

func (r *bookRepository) Update(ctx context.Context, book *model.Book) error {
	if err := r.db.WithContext(ctx).Model(&model.Book{}).Where("id = ?", book.ID).Updates(book).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *bookRepository) Delete(ctx context.Context, id int) error {
	if err := r.db.WithContext(ctx).Delete(&model.Book{}, "id = ?", id).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *bookRepository) List(ctx context.Context) ([]model.Book, error) {
	var books []model.Book
	if err := r.preloaded(ctx).Find(&books).Error; err != nil {
		return nil, common.ErrDBFailed
	}
	return books, nil
}

func (r *bookRepository) GetHighestPrice(ctx context.Context) (*model.Book, error) {
	var book model.Book
	if err := r.preloaded(ctx).Order("price DESC").First(&book).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound
		}
		return nil, common.ErrDBFailed
	}
	return &book, nil
}

func (r *bookRepository) GetMostPurchased(ctx context.Context) (*model.Book, error) {
	var book model.Book
	if err := r.preloaded(ctx).Order("purchase_count DESC").First(&book).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound
		}
		return nil, common.ErrDBFailed
	}
	return &book, nil
}

func (r *bookRepository) IncrementPurchaseCount(ctx context.Context, id int, qty int) error {
	err := r.db.WithContext(ctx).Model(&model.Book{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"purchase_count": gorm.Expr("purchase_count + ?", qty),
			"quantity":       gorm.Expr("quantity - ?", qty),
		}).Error
	if err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *bookRepository) IncrementReadCount(ctx context.Context, id int) error {
	err := r.db.WithContext(ctx).Model(&model.Book{}).Where("id = ?", id).
		UpdateColumn("read_count", gorm.Expr("read_count + ?", 1)).Error
	if err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *bookRepository) PriceStatsByCategory(ctx context.Context) ([]model.CategoryPriceStat, error) {
	var stats []model.CategoryPriceStat
	err := r.db.WithContext(ctx).
		Table("book_store_book AS b").
		Select("b.category_id AS category_id, c.name AS category_name, " +
			"MIN(b.price) AS min_price, MAX(b.price) AS max_price, " +
			"AVG(b.price) AS avg_price, COUNT(*) AS book_count").
		Joins("JOIN book_store_category c ON c.id = b.category_id").
		Group("b.category_id, c.name").
		Scan(&stats).Error
	if err != nil {
		return nil, common.ErrDBFailed
	}
	return stats, nil
}
