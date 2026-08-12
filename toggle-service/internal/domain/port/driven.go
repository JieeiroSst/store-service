package port

import (
	"context"
	"time"

	"github.com/JIeeiroSst/toggle-service/internal/domain/model"
	"github.com/google/uuid"
)

type ProjectRepository interface {
	Create(ctx context.Context, p *model.Project) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Project, error)
	GetByName(ctx context.Context, name string) (*model.Project, error)
	List(ctx context.Context) ([]model.Project, error)
	Update(ctx context.Context, p *model.Project) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type EnvironmentRepository interface {
	Create(ctx context.Context, e *model.Environment) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Environment, error)
	GetByName(ctx context.Context, name string) (*model.Environment, error)
	List(ctx context.Context) ([]model.Environment, error)
	Update(ctx context.Context, e *model.Environment) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type FeatureFlagRepository interface {
	Create(ctx context.Context, f *model.FeatureFlag) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.FeatureFlag, error)
	GetByProjectAndKey(ctx context.Context, projectID uuid.UUID, key string) (*model.FeatureFlag, error)
	ListByProject(ctx context.Context, projectID uuid.UUID) ([]model.FeatureFlag, error)
	Update(ctx context.Context, f *model.FeatureFlag) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type FeatureFlagEnvironmentRepository interface {
	GetOrCreate(ctx context.Context, flagID, environmentID uuid.UUID) (*model.FeatureFlagEnvironment, error)
	GetByFlagAndEnv(ctx context.Context, flagID, environmentID uuid.UUID) (*model.FeatureFlagEnvironment, error)
	GetByFlagAndEnvWithStrategies(ctx context.Context, flagID, environmentID uuid.UUID) (*model.FeatureFlagEnvironment, error)
	SetEnabled(ctx context.Context, id uuid.UUID, enabled bool) error
	ListByEnvironmentWithStrategies(ctx context.Context, projectID, environmentID uuid.UUID) ([]model.FeatureFlagEnvironment, error)
}

type StrategyRepository interface {
	Create(ctx context.Context, s *model.ActivationStrategy) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.ActivationStrategy, error)
	ListByFlagEnvironment(ctx context.Context, flagEnvironmentID uuid.UUID) ([]model.ActivationStrategy, error)
	Update(ctx context.Context, s *model.ActivationStrategy) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type ConstraintRepository interface {
	Create(ctx context.Context, c *model.Constraint) error
	ListByStrategy(ctx context.Context, strategyID uuid.UUID) ([]model.Constraint, error)
	Update(ctx context.Context, c *model.Constraint) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByStrategy(ctx context.Context, strategyID uuid.UUID) error
}

type UserDirectory interface {
	Register(ctx context.Context, email, username, password string) (*model.User, error)
	Login(ctx context.Context, username, password string) (*model.User, error)
}

type RoleRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*model.Role, error)
	GetByName(ctx context.Context, name string) (*model.Role, error)
	List(ctx context.Context) ([]model.Role, error)
	ListPermissions(ctx context.Context, roleID uuid.UUID) ([]model.Permission, error)
}

type MembershipRepository interface {
	Create(ctx context.Context, m *model.ProjectMembership) error
	GetByProjectAndUser(ctx context.Context, projectID uuid.UUID, userID string) (*model.ProjectMembership, error)
	ListByProject(ctx context.Context, projectID uuid.UUID) ([]model.ProjectMembership, error)
	ListByUser(ctx context.Context, userID string) ([]model.ProjectMembership, error)
	UpdateRole(ctx context.Context, id, roleID uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type TokenRepository interface {
	Create(ctx context.Context, t *model.APIToken) error
	GetByHash(ctx context.Context, hash string) (*model.APIToken, error)
	List(ctx context.Context) ([]model.APIToken, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type AuditRepository interface {
	Create(ctx context.Context, e *model.AuditEvent) error
	ListByProject(ctx context.Context, projectID uuid.UUID, entityType string, since, until *time.Time) ([]model.AuditEvent, error)
}

type MetricsRepository interface {
	IncrementCounts(ctx context.Context, flagID, environmentID uuid.UUID, appName string, windowStart, windowStop time.Time, yes, no int64) error
	ListByFlag(ctx context.Context, flagID uuid.UUID) ([]model.FeatureUsageMetric, error)
}
