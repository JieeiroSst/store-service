package repository

import (
	"context"
	"time"

	"github.com/JIeeiroSst/automatic-payment-service/common"
	"github.com/JIeeiroSst/automatic-payment-service/internal/domain/model"
	"github.com/JIeeiroSst/automatic-payment-service/internal/domain/port"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type subscriptionRepo struct {
	db *gorm.DB
}

func NewSubscriptionRepository(db *gorm.DB) port.SubscriptionRepository {
	return &subscriptionRepo{db: db}
}

func (r *subscriptionRepo) Create(ctx context.Context, sub *model.Subscription) error {
	return r.db.WithContext(ctx).Create(sub).Error
}

func (r *subscriptionRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Subscription, error) {
	var sub model.Subscription
	err := r.db.WithContext(ctx).Where("subscription_id = ?", id).First(&sub).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.ErrNotFound
		}
		return nil, common.ErrDBFailed
	}
	return &sub, nil
}

func (r *subscriptionRepo) Update(ctx context.Context, sub *model.Subscription) error {
	return r.db.WithContext(ctx).Save(sub).Error
}

// ListDueForRenewal returns active/trial subscriptions with auto-renewal
// enabled whose next billing date has passed.
func (r *subscriptionRepo) ListDueForRenewal(ctx context.Context, asOf time.Time) ([]model.Subscription, error) {
	var subs []model.Subscription
	err := r.db.WithContext(ctx).
		Where("auto_renewal = ? AND next_billing_date <= ? AND status IN ?",
			true, asOf, []model.SubscriptionStatus{model.SubscriptionActive, model.SubscriptionTrial}).
		Find(&subs).Error
	if err != nil {
		return nil, common.ErrDBFailed
	}
	return subs, nil
}
