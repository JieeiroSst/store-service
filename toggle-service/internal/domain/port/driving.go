package port

import (
	"context"
	"time"

	"github.com/JIeeiroSst/toggle-service/internal/domain/model"
	"github.com/google/uuid"
)

type ProjectService interface {
	Create(ctx context.Context, name, description string, actor string) (*model.Project, error)
	Get(ctx context.Context, id uuid.UUID) (*model.Project, error)
	List(ctx context.Context) ([]model.Project, error)
	Update(ctx context.Context, id uuid.UUID, name, description string, actor string) (*model.Project, error)
	Delete(ctx context.Context, id uuid.UUID, actor string) error
}

type EnvironmentService interface {
	Create(ctx context.Context, name string, envType model.EnvironmentType, sortOrder int) (*model.Environment, error)
	Get(ctx context.Context, id uuid.UUID) (*model.Environment, error)
	List(ctx context.Context) ([]model.Environment, error)
	Update(ctx context.Context, id uuid.UUID, name string, envType model.EnvironmentType, enabled bool) (*model.Environment, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type CreateFlagInput struct {
	ProjectID   uuid.UUID
	Key         string
	Name        string
	Description string
	Type        model.FeatureFlagType
}

type UpdateFlagInput struct {
	Name        string
	Description string
	Type        model.FeatureFlagType
}

type FeatureFlagService interface {
	Create(ctx context.Context, in CreateFlagInput, actor string) (*model.FeatureFlag, error)
	Get(ctx context.Context, projectID uuid.UUID, key string) (*model.FeatureFlag, error)
	List(ctx context.Context, projectID uuid.UUID) ([]model.FeatureFlag, error)
	Update(ctx context.Context, projectID uuid.UUID, key string, in UpdateFlagInput, actor string) (*model.FeatureFlag, error)
	Archive(ctx context.Context, projectID uuid.UUID, key string, actor string) error
	Toggle(ctx context.Context, projectID uuid.UUID, key, environmentName string, enabled bool, actor string) error
}

type StrategyInput struct {
	StrategyType model.StrategyType
	Parameters   map[string]any
	SortOrder    int
	Constraints  []ConstraintInput
}

type ConstraintInput struct {
	ContextField    string
	Operator        model.ConstraintOperator
	Values          []string
	CaseInsensitive bool
}

type StrategyService interface {
	Add(ctx context.Context, projectID uuid.UUID, key, environmentName string, in StrategyInput, actor string) (*model.ActivationStrategy, error)
	Update(ctx context.Context, strategyID uuid.UUID, in StrategyInput, actor string) (*model.ActivationStrategy, error)
	Delete(ctx context.Context, strategyID uuid.UUID, actor string) error
	List(ctx context.Context, projectID uuid.UUID, key, environmentName string) ([]model.ActivationStrategy, error)
}

type AuthService interface {
	Login(ctx context.Context, username, password string) (token string, user *model.User, err error)
	VerifyToken(ctx context.Context, tokenString string) (userID string, isAdmin bool, err error)
	Register(ctx context.Context, email, username, password string) (*model.User, error)
}

type RBACService interface {
	HasPermission(ctx context.Context, userID string, projectID uuid.UUID, permission string) (bool, error)
	AddMember(ctx context.Context, projectID uuid.UUID, userID string, roleID uuid.UUID) (*model.ProjectMembership, error)
	UpdateMemberRole(ctx context.Context, membershipID, roleID uuid.UUID) error
	RemoveMember(ctx context.Context, membershipID uuid.UUID) error
	ListMembers(ctx context.Context, projectID uuid.UUID) ([]model.ProjectMembership, error)
	ListRoles(ctx context.Context) ([]model.Role, error)
}

type CreateTokenInput struct {
	Name          string
	Type          model.APITokenType
	ProjectID     *uuid.UUID
	EnvironmentID *uuid.UUID
	ExpiresAt     *time.Time
}

type TokenService interface {
	Create(ctx context.Context, in CreateTokenInput, actor string) (plaintext string, token *model.APIToken, err error)
	Resolve(ctx context.Context, plaintext string) (*model.APIToken, error)
	List(ctx context.Context) ([]model.APIToken, error)
	Revoke(ctx context.Context, id uuid.UUID) error
}

type AuditService interface {
	Record(ctx context.Context, entityType string, entityID uuid.UUID, action model.AuditAction, projectID, environmentID *uuid.UUID, userID *string, before, after any) error
	List(ctx context.Context, projectID uuid.UUID, entityType string, since, until *time.Time) ([]model.AuditEvent, error)
}

type ClientFeature struct {
	Name        string                     `json:"name"`
	Description string                     `json:"description"`
	Type        model.FeatureFlagType      `json:"type"`
	Enabled     bool                       `json:"enabled"`
	Strategies  []model.ActivationStrategy `json:"strategies"`
}

type ClientFeaturesResponse struct {
	Version  int             `json:"version"`
	Features []ClientFeature `json:"features"`
}

type MetricsBucket struct {
	Start   time.Time                    `json:"start"`
	Stop    time.Time                    `json:"stop"`
	Toggles map[string]MetricsToggleData `json:"toggles"`
}

type MetricsToggleData struct {
	Yes int64 `json:"yes"`
	No  int64 `json:"no"`
}

type MetricsPayload struct {
	AppName string        `json:"appName"`
	Bucket  MetricsBucket `json:"bucket"`
}

type ClientService interface {
	GetFeatures(ctx context.Context, token *model.APIToken) (*ClientFeaturesResponse, error)
	IngestMetrics(ctx context.Context, token *model.APIToken, payload MetricsPayload) error
	Evaluate(ctx context.Context, token *model.APIToken, flagKey string, evalCtx model.EvaluationContext) (bool, error)
}
