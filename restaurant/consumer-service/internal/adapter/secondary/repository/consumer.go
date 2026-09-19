package repository

import (
	"context"

	"github.com/JIeeiroSst/consumer-service/internal/domain/model"
	"github.com/JIeeiroSst/consumer-service/internal/domain/port"
	"github.com/JieeiroSst/logger"
	"gorm.io/gorm"
)

type consumerRepository struct {
	db *gorm.DB
}

func NewConsumerRepository(db *gorm.DB) port.ConsumerRepository {
	return &consumerRepository{db: db}
}

func (r *consumerRepository) Find(ctx context.Context, pagination logger.Pagination) (logger.Pagination, error) {
	var consumers []*model.Consumer

	r.db.WithContext(ctx).Scopes(logger.Paginate(consumers, &pagination, r.db)).Find(&consumers)
	pagination.Rows = consumers

	return pagination, nil
}

func (r *consumerRepository) Create(ctx context.Context, consumer model.Consumer) error {
	return r.db.WithContext(ctx).Create(&consumer).Error
}
