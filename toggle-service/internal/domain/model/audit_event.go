package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type AuditAction string

const (
	ActionCreate    AuditAction = "create"
	ActionUpdate    AuditAction = "update"
	ActionDelete    AuditAction = "delete"
	ActionToggleOn  AuditAction = "toggle_on"
	ActionToggleOff AuditAction = "toggle_off"
)

type AuditEvent struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	EntityType    string         `gorm:"not null;index" json:"entityType"`
	EntityID      uuid.UUID      `gorm:"type:uuid;not null;index" json:"entityId"`
	Action        AuditAction    `gorm:"not null" json:"action"`
	ProjectID     *uuid.UUID     `gorm:"type:uuid;index" json:"projectId,omitempty"`
	EnvironmentID *uuid.UUID     `gorm:"type:uuid" json:"environmentId,omitempty"`
	UserID        *uuid.UUID     `gorm:"type:uuid" json:"userId,omitempty"`
	BeforeJSON    datatypes.JSON `gorm:"type:jsonb" json:"before,omitempty"`
	AfterJSON     datatypes.JSON `gorm:"type:jsonb" json:"after,omitempty"`
	CreatedAt     time.Time      `gorm:"index" json:"createdAt"`
}

func (AuditEvent) TableName() string { return "audit_events" }
