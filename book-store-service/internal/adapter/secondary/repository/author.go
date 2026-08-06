package repository

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/bookStore-service/common"
	"github.com/JIeeiroSst/bookStore-service/internal/domain/model"
	"github.com/JIeeiroSst/bookStore-service/internal/domain/port"
	"gorm.io/gorm"
)

type authorRepository struct {
	db *gorm.DB
}

func NewAuthorRepository(db *gorm.DB) port.AuthorRepository {
	return &authorRepository{db: db}
}

func (r *authorRepository) Create(ctx context.Context, author *model.Author) error {
	if err := r.db.WithContext(ctx).Create(author).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *authorRepository) GetByID(ctx context.Context, id int) (*model.Author, error) {
	var author model.Author
	if err := r.db.WithContext(ctx).First(&author, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound
		}
		return nil, common.ErrDBFailed
	}
	return &author, nil
}

func (r *authorRepository) Update(ctx context.Context, author *model.Author) error {
	if err := r.db.WithContext(ctx).Model(&model.Author{}).Where("id = ?", author.ID).Updates(author).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *authorRepository) Delete(ctx context.Context, id int) error {
	if err := r.db.WithContext(ctx).Delete(&model.Author{}, "id = ?", id).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *authorRepository) List(ctx context.Context) ([]model.Author, error) {
	var authors []model.Author
	if err := r.db.WithContext(ctx).Find(&authors).Error; err != nil {
		return nil, common.ErrDBFailed
	}
	return authors, nil
}

func (r *authorRepository) aggregate(ctx context.Context, orderBy string) (*model.AuthorStat, error) {
	var stat model.AuthorStat
	err := r.db.WithContext(ctx).
		Table("book_store_book AS b").
		Select("b.author_id AS author_id, a.name AS author_name, " +
			"SUM(b.read_count) AS total_reads, SUM(b.purchase_count) AS total_purchases").
		Joins("JOIN book_store_author a ON a.id = b.author_id").
		Group("b.author_id, a.name").
		Order(orderBy + " DESC").
		Limit(1).
		Scan(&stat).Error
	if err != nil {
		return nil, common.ErrDBFailed
	}
	if stat.AuthorID == 0 {
		return nil, common.ErrNotFound
	}
	return &stat, nil
}

func (r *authorRepository) MostRead(ctx context.Context) (*model.AuthorStat, error) {
	return r.aggregate(ctx, "total_reads")
}

func (r *authorRepository) MostPurchased(ctx context.Context) (*model.AuthorStat, error) {
	return r.aggregate(ctx, "total_purchases")
}
