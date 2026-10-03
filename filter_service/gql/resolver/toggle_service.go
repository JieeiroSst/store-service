package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) ToggleQuery() generated.ToggleQueryResolver { return &toggleQueryResolver{r} }

type toggleQueryResolver struct{ *Resolver }

func (r *toggleQueryResolver) Roles(ctx context.Context, obj *model.ToggleQuery) ([]*model.ToggleRole, error) {
	return r.Clients.ToggleService.Roles(ctx)
}

func (r *toggleQueryResolver) Environments(ctx context.Context, obj *model.ToggleQuery) ([]*model.ToggleEnvironment, error) {
	return r.Clients.ToggleService.Environments(ctx)
}

func (r *toggleQueryResolver) Tokens(ctx context.Context, obj *model.ToggleQuery) ([]*model.ToggleAPIToken, error) {
	return r.Clients.ToggleService.Tokens(ctx)
}

func (r *toggleQueryResolver) Projects(ctx context.Context, obj *model.ToggleQuery) ([]*model.ToggleProject, error) {
	return r.Clients.ToggleService.Projects(ctx)
}

func (r *toggleQueryResolver) Project(ctx context.Context, obj *model.ToggleQuery, projectID string) (*model.ToggleProject, error) {
	return r.Clients.ToggleService.Project(ctx, projectID)
}

func (r *toggleQueryResolver) ProjectMembers(ctx context.Context, obj *model.ToggleQuery, projectID string) ([]*model.ToggleProjectMembership, error) {
	return r.Clients.ToggleService.ProjectMembers(ctx, projectID)
}

func (r *toggleQueryResolver) Audit(ctx context.Context, obj *model.ToggleQuery, projectID string, entityType *string, since *string, until *string) ([]*model.ToggleAuditEvent, error) {
	return r.Clients.ToggleService.Audit(ctx, projectID, entityType, since, until)
}

func (r *toggleQueryResolver) Flags(ctx context.Context, obj *model.ToggleQuery, projectID string) ([]*model.ToggleFeatureFlag, error) {
	return r.Clients.ToggleService.Flags(ctx, projectID)
}

func (r *toggleQueryResolver) Flag(ctx context.Context, obj *model.ToggleQuery, projectID string, key string) (*model.ToggleFeatureFlag, error) {
	return r.Clients.ToggleService.Flag(ctx, projectID, key)
}

func (r *toggleQueryResolver) Strategies(ctx context.Context, obj *model.ToggleQuery, projectID string, key string, envName string) ([]*model.ToggleActivationStrategy, error) {
	return r.Clients.ToggleService.Strategies(ctx, projectID, key, envName)
}

func (r *toggleQueryResolver) ClientFeatures(ctx context.Context, obj *model.ToggleQuery) (*model.ToggleClientFeaturesResponse, error) {
	return r.Clients.ToggleService.ClientFeatures(ctx)
}
