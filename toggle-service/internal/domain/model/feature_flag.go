package model

import (
	"time"

	"github.com/google/uuid"
)

type FeatureFlagType string

const (
	FlagTypeRelease    FeatureFlagType = "release"
	FlagTypeExperiment FeatureFlagType = "experiment"
	FlagTypeKillSwitch FeatureFlagType = "kill-switch"
	FlagTypePermission FeatureFlagType = "permission"
)

type FeatureFlag struct {
	ID          uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ProjectID   uuid.UUID       `gorm:"type:uuid;not null;uniqueIndex:idx_project_key" json:"projectId"`
	Key         string          `gorm:"not null;uniqueIndex:idx_project_key" json:"key"`
	Name        string          `gorm:"not null" json:"name"`
	Description string          `json:"description"`
	Type        FeatureFlagType `gorm:"not null;default:release" json:"type"`
	Archived    bool            `gorm:"default:false" json:"archived"`
	CreatedBy   string          `json:"createdBy"` // user_service user ID
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
}

func (FeatureFlag) TableName() string { return "feature_flags" }
