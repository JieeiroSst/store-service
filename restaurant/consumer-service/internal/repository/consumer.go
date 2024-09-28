package repository

import (
	"context"

	"github.com/JIeeiroSst/consumer-service/internal/model"
	"github.com/JieeiroSst/logger"
	"gorm.io/gorm"
)

type Consumers interface {
	Find(ctx context.Context, pagination logger.Pagination) (logger.Pagination, error)
	Create(ctx context.Context, consumer model.Consumer) error
}

type ConsumerRepository struct {
	db *gorm.DB
}

func NewConsumerRepository(db *gorm.DB) *ConsumerRepository {
	return &ConsumerRepository{
		db: db,
	}
}

func (r *ConsumerRepository) Find(ctx context.Context, pagination logger.Pagination) (logger.Pagination, error) {
	var consumers []*model.Consumer

	r.db.Scopes(logger.Paginate(consumers, &pagination, r.db)).Find(&consumers)
	pagination.Rows = consumers

	return pagination, nil
}

func (r *ConsumerRepository) Create(ctx context.Context, consumer model.Consumer) error {
	if err := r.db.Create(&consumer).Error; err != nil {
		return err
	}
	return nil
}
