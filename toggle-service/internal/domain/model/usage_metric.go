package model

import (
	"time"

	"github.com/google/uuid"
)

type FeatureUsageMetric struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	FeatureFlagID uuid.UUID `gorm:"type:uuid;not null;index" json:"featureFlagId"`
	EnvironmentID uuid.UUID `gorm:"type:uuid;not null;index" json:"environmentId"`
	AppName       string    `gorm:"not null" json:"appName"`
	YesCount      int64     `gorm:"default:0" json:"yes"`
	NoCount       int64     `gorm:"default:0" json:"no"`
	WindowStart   time.Time `gorm:"index" json:"windowStart"`
	WindowStop    time.Time `json:"windowStop"`
	CreatedAt     time.Time `json:"createdAt"`
}

func (FeatureUsageMetric) TableName() string { return "feature_usage_metrics" }
