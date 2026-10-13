package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) ToggleCreateEnvironment(ctx context.Context, name string, typeArg *string, sortOrder *int) (*model.ToggleEnvironment, error) {
	return restCall[*model.ToggleEnvironment](ctx, r.Proxy, "toggle-service", rest.Request{Method: "POST", Path: "/api/admin/environments/", Body: map[string]any{"name": name, "type": typeArg, "sortOrder": sortOrder}}, "")
}

func (r *mutationResolver) ToggleUpdateEnvironment(ctx context.Context, environmentID string, name *string, typeArg *string, enabled *bool) (*model.ToggleEnvironment, error) {
	return restCall[*model.ToggleEnvironment](ctx, r.Proxy, "toggle-service", rest.Request{Method: "PUT", Path: rest.P("/api/admin/environments/%v", environmentID), Body: map[string]any{"name": name, "type": typeArg, "enabled": enabled}}, "")
}

func (r *mutationResolver) ToggleDeleteEnvironment(ctx context.Context, environmentID string) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "toggle-service", rest.Request{Method: "DELETE", Path: rest.P("/api/admin/environments/%v", environmentID)}, "")
}

func (r *mutationResolver) ToggleCreateToken(ctx context.Context, name string, typeArg string, projectID *string, environmentID *string, expiresAt *string) (any, error) {
	return restCall[any](ctx, r.Proxy, "toggle-service", rest.Request{Method: "POST", Path: "/api/admin/tokens/", Body: map[string]any{"name": name, "type": typeArg, "projectId": projectID, "environmentId": environmentID, "expiresAt": expiresAt}}, "")
}

func (r *mutationResolver) ToggleRevokeToken(ctx context.Context, tokenID string) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "toggle-service", rest.Request{Method: "DELETE", Path: rest.P("/api/admin/tokens/%v", tokenID)}, "")
}

func (r *mutationResolver) ToggleUpdateMember(ctx context.Context, membershipID string, roleID string) (any, error) {
	return restCall[any](ctx, r.Proxy, "toggle-service", rest.Request{Method: "PUT", Path: rest.P("/api/admin/members/%v/", membershipID), Body: map[string]any{"roleId": roleID}}, "")
}

func (r *mutationResolver) ToggleRemoveMember(ctx context.Context, membershipID string) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "toggle-service", rest.Request{Method: "DELETE", Path: rest.P("/api/admin/members/%v/", membershipID)}, "")
}

func (r *mutationResolver) ToggleCreateProject(ctx context.Context, name string, description *string) (*model.ToggleProject, error) {
	return restCall[*model.ToggleProject](ctx, r.Proxy, "toggle-service", rest.Request{Method: "POST", Path: "/api/admin/projects/", Body: map[string]any{"name": name, "description": description}}, "")
}

func (r *mutationResolver) ToggleUpdateProject(ctx context.Context, projectID string, name *string, description *string) (*model.ToggleProject, error) {
	return restCall[*model.ToggleProject](ctx, r.Proxy, "toggle-service", rest.Request{Method: "PUT", Path: rest.P("/api/admin/projects/%v/", projectID), Body: map[string]any{"name": name, "description": description}}, "")
}

func (r *mutationResolver) ToggleDeleteProject(ctx context.Context, projectID string) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "toggle-service", rest.Request{Method: "DELETE", Path: rest.P("/api/admin/projects/%v/", projectID)}, "")
}

func (r *mutationResolver) ToggleAddMember(ctx context.Context, projectID string, userID string, roleID string) (any, error) {
	return restCall[any](ctx, r.Proxy, "toggle-service", rest.Request{Method: "POST", Path: rest.P("/api/admin/projects/%v/members", projectID), Body: map[string]any{"userId": userID, "roleId": roleID}}, "")
}

func (r *mutationResolver) ToggleCreateFlag(ctx context.Context, projectID string, key string, name string, description *string, typeArg *string) (*model.ToggleFeatureFlag, error) {
	return restCall[*model.ToggleFeatureFlag](ctx, r.Proxy, "toggle-service", rest.Request{Method: "POST", Path: rest.P("/api/admin/projects/%v/flags", projectID), Body: map[string]any{"key": key, "name": name, "description": description, "type": typeArg}}, "")
}

func (r *mutationResolver) ToggleUpdateFlag(ctx context.Context, projectID string, key string, name *string, description *string, typeArg *string) (*model.ToggleFeatureFlag, error) {
	return restCall[*model.ToggleFeatureFlag](ctx, r.Proxy, "toggle-service", rest.Request{Method: "PUT", Path: rest.P("/api/admin/projects/%v/flags/%v/", projectID, key), Body: map[string]any{"name": name, "description": description, "type": typeArg}}, "")
}

func (r *mutationResolver) ToggleArchiveFlag(ctx context.Context, projectID string, key string) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "toggle-service", rest.Request{Method: "DELETE", Path: rest.P("/api/admin/projects/%v/flags/%v/", projectID, key)}, "")
}

func (r *mutationResolver) ToggleFlagOn(ctx context.Context, projectID string, key string, envName string) (any, error) {
	return restCall[any](ctx, r.Proxy, "toggle-service", rest.Request{Method: "PATCH", Path: rest.P("/api/admin/projects/%v/flags/%v/environments/%v/on", projectID, key, envName), Body: map[string]any{}}, "")
}

func (r *mutationResolver) ToggleFlagOff(ctx context.Context, projectID string, key string, envName string) (any, error) {
	return restCall[any](ctx, r.Proxy, "toggle-service", rest.Request{Method: "PATCH", Path: rest.P("/api/admin/projects/%v/flags/%v/environments/%v/off", projectID, key, envName), Body: map[string]any{}}, "")
}

func (r *mutationResolver) ToggleCreateStrategy(ctx context.Context, projectID string, key string, envName string, input model.ToggleStrategyInput) (*model.ToggleActivationStrategy, error) {
	return restCall[*model.ToggleActivationStrategy](ctx, r.Proxy, "toggle-service", rest.Request{Method: "POST", Path: rest.P("/api/admin/projects/%v/flags/%v/environments/%v/strategies", projectID, key, envName), Body: input}, "")
}

func (r *mutationResolver) ToggleUpdateStrategy(ctx context.Context, projectID string, key string, envName string, strategyID string, input model.ToggleStrategyInput) (*model.ToggleActivationStrategy, error) {
	return restCall[*model.ToggleActivationStrategy](ctx, r.Proxy, "toggle-service", rest.Request{Method: "PUT", Path: rest.P("/api/admin/projects/%v/flags/%v/environments/%v/strategies/%v", projectID, key, envName, strategyID), Body: input}, "")
}

func (r *mutationResolver) ToggleDeleteStrategy(ctx context.Context, projectID string, key string, envName string, strategyID string) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "toggle-service", rest.Request{Method: "DELETE", Path: rest.P("/api/admin/projects/%v/flags/%v/environments/%v/strategies/%v", projectID, key, envName, strategyID)}, "")
}

func (r *queryResolver) ToggleRoles(ctx context.Context) ([]*model.ToggleRole, error) {
	return restCall[[]*model.ToggleRole](ctx, r.Proxy, "toggle-service", rest.Request{Method: "GET", Path: "/api/admin/roles"}, "")
}

func (r *queryResolver) ToggleEnvironments(ctx context.Context) ([]*model.ToggleEnvironment, error) {
	return restCall[[]*model.ToggleEnvironment](ctx, r.Proxy, "toggle-service", rest.Request{Method: "GET", Path: "/api/admin/environments/"}, "")
}

func (r *queryResolver) ToggleTokens(ctx context.Context) ([]*model.ToggleAPIToken, error) {
	return restCall[[]*model.ToggleAPIToken](ctx, r.Proxy, "toggle-service", rest.Request{Method: "GET", Path: "/api/admin/tokens/"}, "")
}

func (r *queryResolver) ToggleProjects(ctx context.Context) ([]*model.ToggleProject, error) {
	return restCall[[]*model.ToggleProject](ctx, r.Proxy, "toggle-service", rest.Request{Method: "GET", Path: "/api/admin/projects/"}, "")
}

func (r *queryResolver) ToggleProject(ctx context.Context, projectID string) (*model.ToggleProject, error) {
	return restCall[*model.ToggleProject](ctx, r.Proxy, "toggle-service", rest.Request{Method: "GET", Path: rest.P("/api/admin/projects/%v/", projectID)}, "")
}

func (r *queryResolver) ToggleProjectMembers(ctx context.Context, projectID string) ([]*model.ToggleProjectMembership, error) {
	return restCall[[]*model.ToggleProjectMembership](ctx, r.Proxy, "toggle-service", rest.Request{Method: "GET", Path: rest.P("/api/admin/projects/%v/members", projectID)}, "")
}

func (r *queryResolver) ToggleAudit(ctx context.Context, projectID string, entityType *string, since *string, until *string) ([]*model.ToggleAuditEvent, error) {
	return restCall[[]*model.ToggleAuditEvent](ctx, r.Proxy, "toggle-service", rest.Request{Method: "GET", Path: rest.P("/api/admin/projects/%v/audit", projectID), Query: restQuery("entityType", entityType, "since", since, "until", until)}, "")
}

func (r *queryResolver) ToggleFlags(ctx context.Context, projectID string) ([]*model.ToggleFeatureFlag, error) {
	return restCall[[]*model.ToggleFeatureFlag](ctx, r.Proxy, "toggle-service", rest.Request{Method: "GET", Path: rest.P("/api/admin/projects/%v/flags", projectID)}, "")
}

func (r *queryResolver) ToggleFlag(ctx context.Context, projectID string, key string) (*model.ToggleFeatureFlag, error) {
	return restCall[*model.ToggleFeatureFlag](ctx, r.Proxy, "toggle-service", rest.Request{Method: "GET", Path: rest.P("/api/admin/projects/%v/flags/%v/", projectID, key)}, "")
}

func (r *queryResolver) ToggleStrategies(ctx context.Context, projectID string, key string, envName string) ([]*model.ToggleActivationStrategy, error) {
	return restCall[[]*model.ToggleActivationStrategy](ctx, r.Proxy, "toggle-service", rest.Request{Method: "GET", Path: rest.P("/api/admin/projects/%v/flags/%v/environments/%v/strategies", projectID, key, envName)}, "")
}
