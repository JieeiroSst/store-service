package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type StrategyType string

const (
	StrategyDefault         StrategyType = "default"
	StrategyFlexibleRollout StrategyType = "flexibleRollout"
	StrategyUserWithID      StrategyType = "userWithId"
	StrategyRemoteAddress   StrategyType = "remoteAddress"
)

type ActivationStrategy struct {
	ID                       uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	FeatureFlagEnvironmentID uuid.UUID      `gorm:"type:uuid;not null;index" json:"featureFlagEnvironmentId"`
	StrategyType             StrategyType   `gorm:"not null" json:"strategyType"`
	Parameters               datatypes.JSON `gorm:"type:jsonb" json:"parameters"`
	SortOrder                int            `json:"sortOrder"`
	Constraints              []Constraint   `gorm:"foreignKey:StrategyID" json:"constraints,omitempty"`
	CreatedAt                time.Time      `json:"createdAt"`
	UpdatedAt                time.Time      `json:"updatedAt"`
}

func (ActivationStrategy) TableName() string { return "activation_strategies" }

type FlexibleRolloutParams struct {
	Percentage int    `json:"percentage"`
	Stickiness string `json:"stickiness"` // "userId" | "sessionId" | "random"
}

type UserWithIDParams struct {
	UserIDs []string `json:"userIds"`
}

type RemoteAddressParams struct {
	IPs []string `json:"ips"`
}
