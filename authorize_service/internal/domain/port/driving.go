package port

import (
	"context"

	"github.com/JieeiroSst/authorize-service/internal/domain/model"
	"github.com/JieeiroSst/authorize-service/pkg/pagination"
)

type CasbinUsecase interface {
	Enforce(ctx context.Context, auth model.CasbinAuth) error
	ListRules(ctx context.Context, p pagination.Pagination) (pagination.Pagination, error)
	GetRule(ctx context.Context, id int) (*model.CasbinRule, error)
	CreateRule(ctx context.Context, rule model.CasbinRule) error
	DeleteRule(ctx context.Context, id int) error
	UpdateRuleField(ctx context.Context, id int, field model.UpdateField, value string) error
}

// PermissionUsecase manages user ↔ role assignments and answers
// "what can this user do" questions. user-service goes through it for every
// role change.
type PermissionUsecase interface {
	AssignRoles(ctx context.Context, userID string, roles []string) (model.UserRoles, error)
	RevokeRoles(ctx context.Context, userID string, roles []string) (model.UserRoles, error)
	SetUserRoles(ctx context.Context, userID string, roles []string) (model.UserRoles, error)
	RemoveUser(ctx context.Context, userID string) error
	GetUserRoles(ctx context.Context, userID string) (model.UserRoles, error)
	GetUserPermissions(ctx context.Context, userID string) ([]model.Permission, error)
	CheckPermission(ctx context.Context, userID, obj, act string) (bool, error)
	ListRoles(ctx context.Context) ([]model.RoleDefinition, error)
	// SeedDefaults idempotently writes the built-in role catalog and the
	// bootstrap super admins.
	SeedDefaults(ctx context.Context, superAdminUserIDs []string) error
}

type OTPUsecase interface {
	CreateOtpByUser(ctx context.Context, username string) (string, error)
	Authorize(ctx context.Context, otpCode string, username string) error
}
