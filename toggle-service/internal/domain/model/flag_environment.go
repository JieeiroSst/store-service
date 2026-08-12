package model

import (
	"time"

	"github.com/google/uuid"
)

type FeatureFlagEnvironment struct {
	ID            uuid.UUID            `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	FeatureFlagID uuid.UUID            `gorm:"type:uuid;not null;uniqueIndex:idx_flag_env" json:"featureFlagId"`
	EnvironmentID uuid.UUID            `gorm:"type:uuid;not null;uniqueIndex:idx_flag_env" json:"environmentId"`
	Enabled       bool                 `gorm:"default:false" json:"enabled"`
	Strategies    []ActivationStrategy `gorm:"foreignKey:FeatureFlagEnvironmentID" json:"strategies,omitempty"`
	FeatureFlag   *FeatureFlag         `gorm:"foreignKey:FeatureFlagID" json:"-"`
	UpdatedAt     time.Time            `json:"updatedAt"`
}

func (FeatureFlagEnvironment) TableName() string { return "feature_flag_environments" }
