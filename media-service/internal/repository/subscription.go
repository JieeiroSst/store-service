package repository

import (
	"context"

	"github.com/JIeeiroSst/media-service/model"
	"github.com/JIeeiroSst/utils/logger"
	"gorm.io/gorm"
)

type Subscription interface {
	SaveSubscription(ctx context.Context, subscription model.Subscription) error
}

type SubscriptionRepository struct {
	db *gorm.DB
}

func NewSubscriptionRepository(db *gorm.DB) *SubscriptionRepository {
	return &SubscriptionRepository{
		db: db,
	}
}

func (r *SubscriptionRepository) SaveSubscription(ctx context.Context, subscription model.Subscription) error {
	if err := r.db.Create(&subscription).Error; err != nil {
		logger.Error(ctx, "SaveSubscription error %v", err)
		return err
	}
	return nil
}
