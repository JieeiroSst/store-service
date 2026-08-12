package model

import (
	"time"

	"github.com/google/uuid"
)

type ProjectMembership struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ProjectID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_project_user" json:"projectId"`
	UserID    string    `gorm:"not null;uniqueIndex:idx_project_user" json:"userId"`
	RoleID    uuid.UUID `gorm:"type:uuid;not null" json:"roleId"`
	CreatedAt time.Time `json:"createdAt"`
}

func (ProjectMembership) TableName() string { return "project_memberships" }
