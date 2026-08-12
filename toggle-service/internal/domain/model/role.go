package model

import "github.com/google/uuid"

const (
	RoleOwner  = "Owner"
	RoleMember = "Member"
	RoleViewer = "Viewer"
)

type Role struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name        string    `gorm:"uniqueIndex;not null" json:"name"`
	Description string    `json:"description"`
}

func (Role) TableName() string { return "roles" }

const (
	PermissionCreateFeature  = "CREATE_FEATURE"
	PermissionUpdateFeature  = "UPDATE_FEATURE"
	PermissionDeleteFeature  = "DELETE_FEATURE"
	PermissionToggleFeature  = "TOGGLE_FEATURE"
	PermissionCreateStrategy = "CREATE_STRATEGY"
	PermissionManageProject  = "MANAGE_PROJECT"
	PermissionManageMembers  = "MANAGE_MEMBERS"
	PermissionManageTokens   = "MANAGE_TOKENS"
	PermissionView           = "VIEW"
)

type Permission struct {
	ID   uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name string    `gorm:"uniqueIndex;not null" json:"name"`
}

func (Permission) TableName() string { return "permissions" }

type RolePermission struct {
	RoleID       uuid.UUID `gorm:"type:uuid;primaryKey" json:"roleId"`
	PermissionID uuid.UUID `gorm:"type:uuid;primaryKey" json:"permissionId"`
}

func (RolePermission) TableName() string { return "role_permissions" }
