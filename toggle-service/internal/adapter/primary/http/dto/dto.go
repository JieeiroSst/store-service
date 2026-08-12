package dto

import "github.com/JIeeiroSst/toggle-service/internal/domain/model"

type CreateProjectRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=255"`
	Description string `json:"description"`
}

type UpdateProjectRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=255"`
	Description string `json:"description"`
}

type CreateEnvironmentRequest struct {
	Name      string `json:"name" validate:"required"`
	Type      string `json:"type" validate:"required,oneof=development staging production custom"`
	SortOrder int    `json:"sortOrder"`
}

type UpdateEnvironmentRequest struct {
	Name    string `json:"name" validate:"required"`
	Type    string `json:"type" validate:"required,oneof=development staging production custom"`
	Enabled bool   `json:"enabled"`
}

type CreateFlagRequest struct {
	Key         string `json:"key" validate:"required,min=1,max=255"`
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
	Type        string `json:"type" validate:"omitempty,oneof=release experiment kill-switch permission"`
}

type UpdateFlagRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
	Type        string `json:"type" validate:"omitempty,oneof=release experiment kill-switch permission"`
}

type ConstraintRequest struct {
	ContextField    string   `json:"contextField" validate:"required"`
	Operator        string   `json:"operator" validate:"required,oneof=IN NOT_IN STR_CONTAINS"`
	Values          []string `json:"values"`
	CaseInsensitive bool     `json:"caseInsensitive"`
}

type StrategyRequest struct {
	StrategyType string              `json:"strategyType" validate:"required,oneof=default flexibleRollout userWithId remoteAddress"`
	Parameters   map[string]any      `json:"parameters"`
	SortOrder    int                 `json:"sortOrder"`
	Constraints  []ConstraintRequest `json:"constraints"`
}

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Username string `json:"username" validate:"required,min=3"`
	Password string `json:"password" validate:"required,min=8"`
}

// LoginRequest uses Username (not email) because that's what user_service's
// own login endpoint accepts.
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type AddMemberRequest struct {
	// UserID is a user_service user ID (opaque external identifier, not a
	// local UUID).
	UserID string `json:"userId" validate:"required"`
	RoleID string `json:"roleId" validate:"required,uuid"`
}

type UpdateMemberRequest struct {
	RoleID string `json:"roleId" validate:"required,uuid"`
}

type CreateTokenRequest struct {
	Name          string  `json:"name" validate:"required"`
	Type          string  `json:"type" validate:"required,oneof=client admin"`
	ProjectID     *string `json:"projectId" validate:"omitempty,uuid"`
	EnvironmentID *string `json:"environmentId" validate:"omitempty,uuid"`
	ExpiresAt     *string `json:"expiresAt"` // RFC3339, optional
}

type EvaluateRequest struct {
	FlagKey string                  `json:"flagKey" validate:"required"`
	Context model.EvaluationContext `json:"context"`
}
