package model

import (
	"time"

	"github.com/google/uuid"
)

type SubscriptionStatus string

const (
	SubscriptionActive    SubscriptionStatus = "active"
	SubscriptionTrial     SubscriptionStatus = "trial"
	SubscriptionSuspended SubscriptionStatus = "suspended"
	SubscriptionCancelled SubscriptionStatus = "cancelled"
	SubscriptionExpired   SubscriptionStatus = "expired"
)

type Subscription struct {
	SubscriptionID      uuid.UUID          `gorm:"column:subscription_id;primaryKey" json:"subscriptionId"`
	UserID              uuid.UUID          `gorm:"column:user_id;index;not null" json:"userId"`
	PlanID              uuid.UUID          `gorm:"column:plan_id;not null" json:"planId"`
	Status              SubscriptionStatus `gorm:"column:status;not null" json:"status"`
	Amount              float64            `gorm:"column:amount;not null" json:"amount"`
	Currency            string             `gorm:"column:currency;not null;default:USD" json:"currency"`
	StartDate           time.Time          `gorm:"column:start_date;not null" json:"startDate"`
	EndDate             *time.Time         `gorm:"column:end_date" json:"endDate,omitempty"`
	AutoRenewal         bool               `gorm:"column:auto_renewal;not null;default:true" json:"autoRenewal"`
	TrialEndDate        *time.Time         `gorm:"column:trial_end_date" json:"trialEndDate,omitempty"`
	NextBillingDate     time.Time          `gorm:"column:next_billing_date;index;not null" json:"nextBillingDate"`
	PaymentFailureCount int                `gorm:"column:payment_failure_count;not null;default:0" json:"paymentFailureCount"`
	CreatedAt           time.Time          `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt           time.Time          `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
}

func (Subscription) TableName() string { return "subscriptions" }
