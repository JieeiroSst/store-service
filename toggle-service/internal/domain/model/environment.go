package model

import (
	"time"

	"github.com/google/uuid"
)

type EnvironmentType string

const (
	EnvTypeDevelopment EnvironmentType = "development"
	EnvTypeStaging     EnvironmentType = "staging"
	EnvTypeProduction  EnvironmentType = "production"
	EnvTypeCustom      EnvironmentType = "custom"
)

type Environment struct {
	ID        uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name      string          `gorm:"uniqueIndex;not null" json:"name"`
	Type      EnvironmentType `gorm:"not null" json:"type"`
	SortOrder int             `json:"sortOrder"`
	Enabled   bool            `gorm:"default:true" json:"enabled"`
	CreatedAt time.Time       `json:"createdAt"`
}

func (Environment) TableName() string { return "environments" }
