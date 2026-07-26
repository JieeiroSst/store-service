package repository

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/billing-service/common"
	"github.com/JIeeiroSst/billing-service/internal/domain/model"
	"github.com/JIeeiroSst/billing-service/internal/domain/port"
	"gorm.io/gorm"
)

type subscriptionRepository struct {
	db *gorm.DB
}

func NewSubscriptionRepository(db *gorm.DB) port.SubscriptionRepository {
	return &subscriptionRepository{db: db}
}

func (r *subscriptionRepository) Create(ctx context.Context, subscription *model.Subscription) error {
	if err := r.db.WithContext(ctx).Create(subscription).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *subscriptionRepository) GetByID(ctx context.Context, id int) (*model.Subscription, error) {
	var subscription model.Subscription
	if err := r.db.WithContext(ctx).Where("subscription_id = ?", id).First(&subscription).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound
		}
		return nil, common.ErrDBFailed
	}
	return &subscription, nil
}

func (r *subscriptionRepository) Update(ctx context.Context, subscription *model.Subscription) error {
	if err := r.db.WithContext(ctx).Model(&model.Subscription{}).
		Where("subscription_id = ?", subscription.SubscriptionID).
		Updates(subscription).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *subscriptionRepository) Delete(ctx context.Context, id int) error {
	if err := r.db.WithContext(ctx).Where("subscription_id = ?", id).Delete(&model.Subscription{}).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *subscriptionRepository) List(ctx context.Context) ([]model.Subscription, error) {
	var subscriptions []model.Subscription
	if err := r.db.WithContext(ctx).Find(&subscriptions).Error; err != nil {
		return nil, common.ErrDBFailed
	}
	return subscriptions, nil
}
