package model

import (
	"time"

	"github.com/google/uuid"
)

type APITokenType string

const (
	TokenTypeClient APITokenType = "client"
	TokenTypeAdmin  APITokenType = "admin"
)

type APIToken struct {
	ID            uuid.UUID    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name          string       `gorm:"not null" json:"name"`
	TokenHash     string       `gorm:"uniqueIndex;not null" json:"-"`
	TokenPrefix   string       `gorm:"not null" json:"tokenPrefix"`
	Type          APITokenType `gorm:"not null" json:"type"`
	ProjectID     *uuid.UUID   `gorm:"type:uuid" json:"projectId,omitempty"`
	EnvironmentID *uuid.UUID   `gorm:"type:uuid" json:"environmentId,omitempty"`
	ExpiresAt     *time.Time   `json:"expiresAt,omitempty"`
	CreatedBy     string       `json:"createdBy"` // user_service user ID
	CreatedAt     time.Time    `json:"createdAt"`
}

func (APIToken) TableName() string { return "api_tokens" }
