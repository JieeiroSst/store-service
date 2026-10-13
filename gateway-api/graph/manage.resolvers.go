package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) KcPostRealmsProtocolToken(ctx context.Context, realm string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/realms/%v/protocol/token", realm), Body: input}, "")
}

func (r *mutationResolver) KcPostRealmsProtocolIntrospect(ctx context.Context, realm string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/realms/%v/protocol/introspect", realm), Body: input}, "")
}

func (r *mutationResolver) KcPostRealmsProtocolRevoke(ctx context.Context, realm string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/realms/%v/protocol/revoke", realm), Body: input}, "")
}

func (r *mutationResolver) KcPostRealmsProtocolTokenExchange(ctx context.Context, realm string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/realms/%v/protocol/token-exchange", realm), Body: input}, "")
}

func (r *mutationResolver) KcPostRealmsProtocolLogoutPublic(ctx context.Context, realm string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/realms/%v/protocol/logout-public", realm), Body: input}, "")
}

func (r *mutationResolver) KcPostRealmsAuthzRpt(ctx context.Context, realm string) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/realms/%v/authz/rpt", realm), Body: map[string]any{}}, "")
}

func (r *mutationResolver) KcPostRealmsAuthzRptPermissions(ctx context.Context, realm string) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/realms/%v/authz/rpt/permissions", realm), Body: map[string]any{}}, "")
}

func (r *mutationResolver) KcPostRealmsAuthzRptDecision(ctx context.Context, realm string) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/realms/%v/authz/rpt/decision", realm), Body: map[string]any{}}, "")
}

func (r *mutationResolver) KcPostRealmsAuthzPermissionTickets(ctx context.Context, realm string, input []map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/realms/%v/authz/permission-tickets", realm), Body: input}, "")
}

func (r *mutationResolver) KcPostRealmsAuthzUmaPermissions(ctx context.Context, realm string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/realms/%v/authz/uma-permissions", realm), Body: input}, "")
}

func (r *mutationResolver) KcPutRealmsAuthzUmaPermissions(ctx context.Context, realm string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/realms/%v/authz/uma-permissions", realm), Body: input}, "")
}

func (r *mutationResolver) KcDeleteRealmsAuthzUmaPermissions(ctx context.Context, realm string, ticketID string) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/realms/%v/authz/uma-permissions/%v", realm, ticketID)}, "")
}

func (r *mutationResolver) KcPostRealmsAuthzResources(ctx context.Context, realm string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/realms/%v/authz/resources", realm), Body: input}, "")
}

func (r *mutationResolver) KcPutRealmsAuthzResources(ctx context.Context, realm string, resourceID string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/realms/%v/authz/resources/%v", realm, resourceID), Body: input}, "")
}

func (r *mutationResolver) KcDeleteRealmsAuthzResources(ctx context.Context, realm string, resourceID string) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/realms/%v/authz/resources/%v", realm, resourceID)}, "")
}

func (r *mutationResolver) KcPostRealmsAuthzResourcesPolicies(ctx context.Context, realm string, resourceID string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/realms/%v/authz/resources/%v/policies", realm, resourceID), Body: input}, "")
}

func (r *mutationResolver) KcPutRealmsAuthzResourcePolicies(ctx context.Context, realm string, permissionID string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/realms/%v/authz/resource-policies/%v", realm, permissionID), Body: input}, "")
}

func (r *mutationResolver) KcDeleteRealmsAuthzResourcePolicies(ctx context.Context, realm string, permissionID string) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/realms/%v/authz/resource-policies/%v", realm, permissionID)}, "")
}

func (r *mutationResolver) KcPostAdminRealms(ctx context.Context, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "POST", Path: "/api/v1/admin/realms", Body: input}, "")
}

func (r *mutationResolver) KcPutAdminRealms(ctx context.Context, realm string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/admin/realms/%v", realm), Body: input}, "")
}

func (r *mutationResolver) KcDeleteAdminRealms(ctx context.Context, realm string) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/admin/realms/%v", realm)}, "")
}

func (r *mutationResolver) KcPostAdminRealmsClearRealmCache(ctx context.Context, realm string) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/admin/realms/%v/clear-realm-cache", realm), Body: map[string]any{}}, "")
}

func (r *mutationResolver) KcPostAdminRealmsClearUserCache(ctx context.Context, realm string) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/admin/realms/%v/clear-user-cache", realm), Body: map[string]any{}}, "")
}

func (r *mutationResolver) KcPostAdminRealmsClearKeysCache(ctx context.Context, realm string) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/admin/realms/%v/clear-keys-cache", realm), Body: map[string]any{}}, "")
}

func (r *mutationResolver) KcDeleteAdminRealmsSessions(ctx context.Context, realm string, sessionID string) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/admin/realms/%v/sessions/%v", realm, sessionID)}, "")
}

func (r *mutationResolver) KcPostAdminRealmsUsers(ctx context.Context, realm string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/admin/realms/%v/users", realm), Body: input}, "")
}

func (r *mutationResolver) KcPutAdminRealmsUsers(ctx context.Context, realm string, userID string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/admin/realms/%v/users/%v", realm, userID), Body: input}, "")
}

func (r *mutationResolver) KcDeleteAdminRealmsUsers(ctx context.Context, realm string, userID string) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/admin/realms/%v/users/%v", realm, userID)}, "")
}

func (r *mutationResolver) KcPutAdminRealmsUsersResetPassword(ctx context.Context, realm string, userID string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/admin/realms/%v/users/%v/reset-password", realm, userID), Body: input}, "")
}

func (r *mutationResolver) KcPutAdminRealmsUsersExecuteActionsEmail(ctx context.Context, realm string, userID string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/admin/realms/%v/users/%v/execute-actions-email", realm, userID), Body: input}, "")
}

func (r *mutationResolver) KcPutAdminRealmsUsersSendVerifyEmail(ctx context.Context, realm string, userID string) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/admin/realms/%v/users/%v/send-verify-email", realm, userID), Body: map[string]any{}}, "")
}

func (r *mutationResolver) KcDeleteAdminRealmsUsersConsents(ctx context.Context, realm string, userID string, clientID string) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/admin/realms/%v/users/%v/consents/%v", realm, userID, clientID)}, "")
}

func (r *mutationResolver) KcPostAdminRealmsUsersLogout(ctx context.Context, realm string, userID string) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/admin/realms/%v/users/%v/logout", realm, userID), Body: map[string]any{}}, "")
}

func (r *mutationResolver) KcPutAdminRealmsUsersGroups(ctx context.Context, realm string, userID string, groupID string) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/admin/realms/%v/users/%v/groups/%v", realm, userID, groupID), Body: map[string]any{}}, "")
}

func (r *mutationResolver) KcDeleteAdminRealmsUsersGroups(ctx context.Context, realm string, userID string, groupID string) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/admin/realms/%v/users/%v/groups/%v", realm, userID, groupID)}, "")
}

func (r *mutationResolver) KcDeleteAdminRealmsUsersCredentials(ctx context.Context, realm string, userID string, credentialID string) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/admin/realms/%v/users/%v/credentials/%v", realm, userID, credentialID)}, "")
}

func (r *mutationResolver) KcPutAdminRealmsUsersCredentialsUserLabel(ctx context.Context, realm string, userID string, credentialID string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/admin/realms/%v/users/%v/credentials/%v/userLabel", realm, userID, credentialID), Body: input}, "")
}

func (r *mutationResolver) KcPostAdminRealmsUsersCredentialsMoveToFirst(ctx context.Context, realm string, userID string, credentialID string) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/admin/realms/%v/users/%v/credentials/%v/moveToFirst", realm, userID, credentialID), Body: map[string]any{}}, "")
}

func (r *mutationResolver) KcPostAdminRealmsUsersCredentialsMoveAfter(ctx context.Context, realm string, userID string, credentialID string, previousCredentialID string) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/admin/realms/%v/users/%v/credentials/%v/moveAfter/%v", realm, userID, credentialID, previousCredentialID), Body: map[string]any{}}, "")
}

func (r *mutationResolver) KcPutAdminRealmsUsersDisableCredentialTypes(ctx context.Context, realm string, userID string, input []string) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/admin/realms/%v/users/%v/disable-credential-types", realm, userID), Body: input}, "")
}

func (r *mutationResolver) KcPostAdminRealmsUsersFederatedIdentity(ctx context.Context, realm string, userID string, providerID string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/admin/realms/%v/users/%v/federated-identity/%v", realm, userID, providerID), Body: input}, "")
}

func (r *mutationResolver) KcDeleteAdminRealmsUsersFederatedIdentity(ctx context.Context, realm string, userID string, providerID string) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/admin/realms/%v/users/%v/federated-identity/%v", realm, userID, providerID)}, "")
}

func (r *mutationResolver) KcPostAdminRealmsGroups(ctx context.Context, realm string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/admin/realms/%v/groups", realm), Body: input}, "")
}

func (r *mutationResolver) KcPutAdminRealmsGroups(ctx context.Context, realm string, groupID string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/admin/realms/%v/groups/%v", realm, groupID), Body: input}, "")
}

func (r *mutationResolver) KcDeleteAdminRealmsGroups(ctx context.Context, realm string, groupID string) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/admin/realms/%v/groups/%v", realm, groupID)}, "")
}

func (r *mutationResolver) KcPostAdminRealmsGroupsChildren(ctx context.Context, realm string, groupID string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/admin/realms/%v/groups/%v/children", realm, groupID), Body: input}, "")
}

func (r *mutationResolver) KcPutAdminRealmsDefaultGroups(ctx context.Context, realm string, groupID string) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/admin/realms/%v/default-groups/%v", realm, groupID), Body: map[string]any{}}, "")
}

func (r *mutationResolver) KcDeleteAdminRealmsDefaultGroups(ctx context.Context, realm string, groupID string) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/admin/realms/%v/default-groups/%v", realm, groupID)}, "")
}

func (r *mutationResolver) KcPostAdminRealmsRoles(ctx context.Context, realm string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/admin/realms/%v/roles", realm), Body: input}, "")
}

func (r *mutationResolver) KcPutAdminRealmsRoles(ctx context.Context, realm string, roleName string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/admin/realms/%v/roles/%v", realm, roleName), Body: input}, "")
}

func (r *mutationResolver) KcDeleteAdminRealmsRoles(ctx context.Context, realm string, roleName string) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/admin/realms/%v/roles/%v", realm, roleName)}, "")
}

func (r *mutationResolver) KcPostAdminRealmsRolesComposites(ctx context.Context, realm string, roleName string, input []map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/admin/realms/%v/roles/%v/composites", realm, roleName), Body: input}, "")
}

func (r *mutationResolver) KcDeleteAdminRealmsRolesComposites(ctx context.Context, realm string, roleName string, input []map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/admin/realms/%v/roles/%v/composites", realm, roleName), Body: input}, "")
}

func (r *mutationResolver) KcPutAdminRealmsRolesByID(ctx context.Context, realm string, roleID string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/admin/realms/%v/roles-by-id/%v", realm, roleID), Body: input}, "")
}

func (r *mutationResolver) KcPostAdminRealmsRolesByIDComposites(ctx context.Context, realm string, roleID string, input []map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/admin/realms/%v/roles-by-id/%v/composites", realm, roleID), Body: input}, "")
}

func (r *mutationResolver) KcDeleteAdminRealmsRolesByIDComposites(ctx context.Context, realm string, roleID string, input []map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/admin/realms/%v/roles-by-id/%v/composites", realm, roleID), Body: input}, "")
}

func (r *mutationResolver) KcPostAdminRealmsClients(ctx context.Context, realm string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/admin/realms/%v/clients", realm), Body: input}, "")
}

func (r *mutationResolver) KcPutAdminRealmsClients(ctx context.Context, realm string, clientID string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/admin/realms/%v/clients/%v", realm, clientID), Body: input}, "")
}

func (r *mutationResolver) KcDeleteAdminRealmsClients(ctx context.Context, realm string, clientID string) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/admin/realms/%v/clients/%v", realm, clientID)}, "")
}

func (r *mutationResolver) KcPostAdminRealmsClientsClientSecret(ctx context.Context, realm string, clientID string) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/admin/realms/%v/clients/%v/client-secret", realm, clientID), Body: map[string]any{}}, "")
}

func (r *mutationResolver) KcPutAdminRealmsClientsDefaultClientScopes(ctx context.Context, realm string, clientID string, scopeID string) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/admin/realms/%v/clients/%v/default-client-scopes/%v", realm, clientID, scopeID), Body: map[string]any{}}, "")
}

func (r *mutationResolver) KcDeleteAdminRealmsClientsDefaultClientScopes(ctx context.Context, realm string, clientID string, scopeID string) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/admin/realms/%v/clients/%v/default-client-scopes/%v", realm, clientID, scopeID)}, "")
}

func (r *mutationResolver) KcPutAdminRealmsClientsOptionalClientScopes(ctx context.Context, realm string, clientID string, scopeID string) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/admin/realms/%v/clients/%v/optional-client-scopes/%v", realm, clientID, scopeID), Body: map[string]any{}}, "")
}

func (r *mutationResolver) KcDeleteAdminRealmsClientsOptionalClientScopes(ctx context.Context, realm string, clientID string, scopeID string) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/admin/realms/%v/clients/%v/optional-client-scopes/%v", realm, clientID, scopeID)}, "")
}

func (r *mutationResolver) KcPostAdminRealmsClientsRoles(ctx context.Context, realm string, clientID string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/admin/realms/%v/clients/%v/roles", realm, clientID), Body: input}, "")
}

func (r *mutationResolver) KcPutAdminRealmsClientsRoles(ctx context.Context, realm string, clientID string, roleName string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/admin/realms/%v/clients/%v/roles/%v", realm, clientID, roleName), Body: input}, "")
}

func (r *mutationResolver) KcDeleteAdminRealmsClientsRoles(ctx context.Context, realm string, clientID string, roleName string) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/admin/realms/%v/clients/%v/roles/%v", realm, clientID, roleName)}, "")
}

func (r *mutationResolver) KcPostAdminRealmsClientsProtocolMappers(ctx context.Context, realm string, clientID string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/admin/realms/%v/clients/%v/protocol-mappers", realm, clientID), Body: input}, "")
}

func (r *mutationResolver) KcPutAdminRealmsClientsProtocolMappers(ctx context.Context, realm string, clientID string, mapperID string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/admin/realms/%v/clients/%v/protocol-mappers/%v", realm, clientID, mapperID), Body: input}, "")
}

func (r *mutationResolver) KcDeleteAdminRealmsClientsProtocolMappers(ctx context.Context, realm string, clientID string, mapperID string) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/admin/realms/%v/clients/%v/protocol-mappers/%v", realm, clientID, mapperID)}, "")
}

func (r *mutationResolver) KcPostAdminRealmsClientsAuthzResourceServerResource(ctx context.Context, realm string, clientID string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/admin/realms/%v/clients/%v/authz/resource-server/resource", realm, clientID), Body: input}, "")
}

func (r *mutationResolver) KcPutAdminRealmsClientsAuthzResourceServerResource(ctx context.Context, realm string, clientID string, resourceID string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/admin/realms/%v/clients/%v/authz/resource-server/resource/%v", realm, clientID, resourceID), Body: input}, "")
}

func (r *mutationResolver) KcDeleteAdminRealmsClientsAuthzResourceServerResource(ctx context.Context, realm string, clientID string, resourceID string) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/admin/realms/%v/clients/%v/authz/resource-server/resource/%v", realm, clientID, resourceID)}, "")
}

func (r *mutationResolver) KcPostAdminRealmsClientsAuthzResourceServerScope(ctx context.Context, realm string, clientID string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/admin/realms/%v/clients/%v/authz/resource-server/scope", realm, clientID), Body: input}, "")
}

func (r *mutationResolver) KcPutAdminRealmsClientsAuthzResourceServerScope(ctx context.Context, realm string, clientID string, scopeID string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/admin/realms/%v/clients/%v/authz/resource-server/scope/%v", realm, clientID, scopeID), Body: input}, "")
}

func (r *mutationResolver) KcDeleteAdminRealmsClientsAuthzResourceServerScope(ctx context.Context, realm string, clientID string, scopeID string) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/admin/realms/%v/clients/%v/authz/resource-server/scope/%v", realm, clientID, scopeID)}, "")
}

func (r *mutationResolver) KcPostAdminRealmsClientsAuthzResourceServerPolicy(ctx context.Context, realm string, clientID string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/admin/realms/%v/clients/%v/authz/resource-server/policy", realm, clientID), Body: input}, "")
}

func (r *mutationResolver) KcPutAdminRealmsClientsAuthzResourceServerPolicy(ctx context.Context, realm string, clientID string, policyID string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/admin/realms/%v/clients/%v/authz/resource-server/policy/%v", realm, clientID, policyID), Body: input}, "")
}

func (r *mutationResolver) KcDeleteAdminRealmsClientsAuthzResourceServerPolicy(ctx context.Context, realm string, clientID string, policyID string) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/admin/realms/%v/clients/%v/authz/resource-server/policy/%v", realm, clientID, policyID)}, "")
}

func (r *mutationResolver) KcPostAdminRealmsClientsAuthzResourceServerPermission(ctx context.Context, realm string, clientID string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/admin/realms/%v/clients/%v/authz/resource-server/permission", realm, clientID), Body: input}, "")
}

func (r *mutationResolver) KcPutAdminRealmsClientsAuthzResourceServerPermission(ctx context.Context, realm string, clientID string, permissionID string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/admin/realms/%v/clients/%v/authz/resource-server/permission/%v", realm, clientID, permissionID), Body: input}, "")
}

func (r *mutationResolver) KcDeleteAdminRealmsClientsAuthzResourceServerPermission(ctx context.Context, realm string, clientID string, permissionID string) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/admin/realms/%v/clients/%v/authz/resource-server/permission/%v", realm, clientID, permissionID)}, "")
}

func (r *mutationResolver) KcPostAdminRealmsClientsRegistrations(ctx context.Context, realm string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/admin/realms/%v/clients-registrations", realm), Body: input}, "")
}

func (r *mutationResolver) KcPutAdminRealmsClientsRegistrations(ctx context.Context, realm string, clientID string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/admin/realms/%v/clients-registrations/%v", realm, clientID), Body: input}, "")
}

func (r *mutationResolver) KcDeleteAdminRealmsClientsRegistrations(ctx context.Context, realm string, clientID string) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/admin/realms/%v/clients-registrations/%v", realm, clientID)}, "")
}

func (r *mutationResolver) KcPostAdminRealmsClientScopes(ctx context.Context, realm string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/admin/realms/%v/client-scopes", realm), Body: input}, "")
}

func (r *mutationResolver) KcPutAdminRealmsClientScopes(ctx context.Context, realm string, scopeID string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/admin/realms/%v/client-scopes/%v", realm, scopeID), Body: input}, "")
}

func (r *mutationResolver) KcDeleteAdminRealmsClientScopes(ctx context.Context, realm string, scopeID string) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/admin/realms/%v/client-scopes/%v", realm, scopeID)}, "")
}

func (r *mutationResolver) KcPostAdminRealmsClientScopesProtocolMappers(ctx context.Context, realm string, scopeID string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/admin/realms/%v/client-scopes/%v/protocol-mappers", realm, scopeID), Body: input}, "")
}

func (r *mutationResolver) KcPutAdminRealmsClientScopesProtocolMappers(ctx context.Context, realm string, scopeID string, mapperID string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/admin/realms/%v/client-scopes/%v/protocol-mappers/%v", realm, scopeID, mapperID), Body: input}, "")
}

func (r *mutationResolver) KcDeleteAdminRealmsClientScopesProtocolMappers(ctx context.Context, realm string, scopeID string, mapperID string) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/admin/realms/%v/client-scopes/%v/protocol-mappers/%v", realm, scopeID, mapperID)}, "")
}

func (r *mutationResolver) KcPostAdminRealmsComponents(ctx context.Context, realm string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/admin/realms/%v/components", realm), Body: input}, "")
}

func (r *mutationResolver) KcPutAdminRealmsComponents(ctx context.Context, realm string, componentID string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/admin/realms/%v/components/%v", realm, componentID), Body: input}, "")
}

func (r *mutationResolver) KcDeleteAdminRealmsComponents(ctx context.Context, realm string, componentID string) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/admin/realms/%v/components/%v", realm, componentID)}, "")
}

func (r *mutationResolver) KcPostAdminRealmsAuthenticationFlows(ctx context.Context, realm string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/admin/realms/%v/authentication/flows", realm), Body: input}, "")
}

func (r *mutationResolver) KcPutAdminRealmsAuthenticationFlows(ctx context.Context, realm string, flow string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/admin/realms/%v/authentication/flows/%v", realm, flow), Body: input}, "")
}

func (r *mutationResolver) KcDeleteAdminRealmsAuthenticationFlows(ctx context.Context, realm string, flow string) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/admin/realms/%v/authentication/flows/%v", realm, flow)}, "")
}

func (r *mutationResolver) KcPutAdminRealmsAuthenticationFlowsExecutions(ctx context.Context, realm string, flow string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/admin/realms/%v/authentication/flows/%v/executions", realm, flow), Body: input}, "")
}

func (r *mutationResolver) KcPostAdminRealmsAuthenticationFlowsExecutionsExecution(ctx context.Context, realm string, flow string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/admin/realms/%v/authentication/flows/%v/executions/execution", realm, flow), Body: input}, "")
}

func (r *mutationResolver) KcPostAdminRealmsAuthenticationFlowsExecutionsFlow(ctx context.Context, realm string, flow string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/admin/realms/%v/authentication/flows/%v/executions/flow", realm, flow), Body: input}, "")
}

func (r *mutationResolver) KcDeleteAdminRealmsAuthenticationExecutions(ctx context.Context, realm string, executionID string) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/admin/realms/%v/authentication/executions/%v", realm, executionID)}, "")
}

func (r *mutationResolver) KcPostAdminRealmsAuthenticationRegisterRequiredAction(ctx context.Context, realm string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/admin/realms/%v/authentication/register-required-action", realm), Body: input}, "")
}

func (r *mutationResolver) KcPutAdminRealmsAuthenticationRequiredActions(ctx context.Context, realm string, alias string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/admin/realms/%v/authentication/required-actions/%v", realm, alias), Body: input}, "")
}

func (r *mutationResolver) KcDeleteAdminRealmsAuthenticationRequiredActions(ctx context.Context, realm string, alias string) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/admin/realms/%v/authentication/required-actions/%v", realm, alias)}, "")
}

func (r *mutationResolver) KcPostAdminRealmsIdentityProviders(ctx context.Context, realm string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/admin/realms/%v/identity-providers", realm), Body: input}, "")
}

func (r *mutationResolver) KcPostAdminRealmsIdentityProvidersImportConfig(ctx context.Context, realm string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/admin/realms/%v/identity-providers/import-config", realm), Body: input}, "")
}

func (r *mutationResolver) KcPutAdminRealmsIdentityProviders(ctx context.Context, realm string, alias string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/admin/realms/%v/identity-providers/%v", realm, alias), Body: input}, "")
}

func (r *mutationResolver) KcDeleteAdminRealmsIdentityProviders(ctx context.Context, realm string, alias string) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/admin/realms/%v/identity-providers/%v", realm, alias)}, "")
}

func (r *mutationResolver) KcPostAdminRealmsIdentityProvidersMappers(ctx context.Context, realm string, alias string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/admin/realms/%v/identity-providers/%v/mappers", realm, alias), Body: input}, "")
}

func (r *mutationResolver) KcPutAdminRealmsIdentityProvidersMappers(ctx context.Context, realm string, alias string, mapperID string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/admin/realms/%v/identity-providers/%v/mappers/%v", realm, alias, mapperID), Body: input}, "")
}

func (r *mutationResolver) KcDeleteAdminRealmsIdentityProvidersMappers(ctx context.Context, realm string, alias string, mapperID string) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/admin/realms/%v/identity-providers/%v/mappers/%v", realm, alias, mapperID)}, "")
}

func (r *mutationResolver) KcPostLoginAdmin(ctx context.Context, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "POST", Path: "/login-admin", Body: input}, "")
}

func (r *mutationResolver) KcPostUser(ctx context.Context, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "POST", Path: "/user", Body: input}, "")
}

func (r *mutationResolver) KcPostLogin(ctx context.Context, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "POST", Path: "/login", Body: input}, "")
}

func (r *mutationResolver) KcPostLoginOtp(ctx context.Context, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "POST", Path: "/login-otp", Body: input}, "")
}

func (r *mutationResolver) KcPostLogout(ctx context.Context, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "POST", Path: "/logout", Body: input}, "")
}

func (r *mutationResolver) KcPostLoginClient(ctx context.Context, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "POST", Path: "/login-client", Body: input}, "")
}

func (r *mutationResolver) KcPostRefreshToken(ctx context.Context, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "POST", Path: "/refresh-token", Body: input}, "")
}

func (r *mutationResolver) KcPostUserInfo(ctx context.Context, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "POST", Path: "/user-info", Body: input}, "")
}

func (r *mutationResolver) KcPostSetPassword(ctx context.Context, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "manage-service", rest.Request{Method: "POST", Path: "/set-password", Body: input}, "")
}

func (r *queryResolver) KcTokenUser(ctx context.Context, realm *string) (*model.KcTokenInfo, error) {
	return restCall[*model.KcTokenInfo](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: "/token-user", Query: restQuery("realm", realm)}, "")
}

func (r *queryResolver) KcIssuer(ctx context.Context, realm string) (*model.KcIssuerResponse, error) {
	return restCall[*model.KcIssuerResponse](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/realms/%v/issuer", realm)}, "")
}

func (r *queryResolver) KcCerts(ctx context.Context, realm string) (*model.KcCertResponse, error) {
	return restCall[*model.KcCertResponse](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/realms/%v/protocol/certs", realm)}, "")
}

func (r *queryResolver) KcRawUserInfo(ctx context.Context, realm string) (map[string]any, error) {
	return restCall[map[string]any](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/realms/%v/protocol/userinfo", realm)}, "")
}

func (r *queryResolver) KcDecodeAccessToken(ctx context.Context, realm string) (map[string]any, error) {
	return restCall[map[string]any](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/realms/%v/protocol/decode", realm)}, "")
}

func (r *queryResolver) KcUserPermissions(ctx context.Context, realm string) ([]*model.KcPermissionGrantResponseRepresentation, error) {
	return restCall[[]*model.KcPermissionGrantResponseRepresentation](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/realms/%v/authz/uma-permissions", realm)}, "")
}

func (r *queryResolver) KcResourcesClient(ctx context.Context, realm string, deep *string, first *string, max *string, name *string, owner *string, typeArg *string, uri *string, scope *string, matchingURI *string, exactName *string) ([]*model.KcResourceRepresentation, error) {
	return restCall[[]*model.KcResourceRepresentation](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/realms/%v/authz/resources", realm), Query: restQuery("deep", deep, "first", first, "max", max, "name", name, "owner", owner, "type", typeArg, "uri", uri, "scope", scope, "matchingUri", matchingURI, "exactName", exactName)}, "")
}

func (r *queryResolver) KcResourceClient(ctx context.Context, realm string, resourceID string) (*model.KcResourceRepresentation, error) {
	return restCall[*model.KcResourceRepresentation](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/realms/%v/authz/resources/%v", realm, resourceID)}, "")
}

func (r *queryResolver) KcResourcePolicies(ctx context.Context, realm string, resource *string, name *string, scope *string, first *string, max *string) ([]*model.KcResourcePolicyRepresentation, error) {
	return restCall[[]*model.KcResourcePolicyRepresentation](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/realms/%v/authz/resource-policies", realm), Query: restQuery("resource", resource, "name", name, "scope", scope, "first", first, "max", max)}, "")
}

func (r *queryResolver) KcResourcePolicy(ctx context.Context, realm string, permissionID string) (*model.KcResourcePolicyRepresentation, error) {
	return restCall[*model.KcResourcePolicyRepresentation](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/realms/%v/authz/resource-policies/%v", realm, permissionID)}, "")
}

func (r *queryResolver) KcServerInfo(ctx context.Context) (*model.KcServerInfoRepresentation, error) {
	return restCall[*model.KcServerInfoRepresentation](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: "/api/v1/admin/server-info"}, "")
}

func (r *queryResolver) KcRealms(ctx context.Context) ([]*model.KcRealmRepresentation, error) {
	return restCall[[]*model.KcRealmRepresentation](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: "/api/v1/admin/realms"}, "")
}

func (r *queryResolver) KcRealm(ctx context.Context, realm string) (*model.KcRealmRepresentation, error) {
	return restCall[*model.KcRealmRepresentation](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v", realm)}, "")
}

func (r *queryResolver) KcKeyStoreConfig(ctx context.Context, realm string) (*model.KcKeyStoreConfig, error) {
	return restCall[*model.KcKeyStoreConfig](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/keys", realm)}, "")
}

func (r *queryResolver) KcEvents(ctx context.Context, realm string, client *string, dateFrom *string, dateTo *string, first *string, ipAddress *string, max *string, user *string) ([]*model.KcEventRepresentation, error) {
	return restCall[[]*model.KcEventRepresentation](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/events", realm), Query: restQuery("client", client, "dateFrom", dateFrom, "dateTo", dateTo, "first", first, "ipAddress", ipAddress, "max", max, "user", user)}, "")
}

func (r *queryResolver) KcCredentialRegistrators(ctx context.Context, realm string) ([]string, error) {
	return restCall[[]string](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/credential-registrators", realm)}, "")
}

func (r *queryResolver) KcUsers(ctx context.Context, realm string, briefRepresentation *string, email *string, emailVerified *string, enabled *string, exact *string, first *string, firstName *string, idpAlias *string, idpUserID *string, lastName *string, max *string, q *string, search *string, username *string) ([]*model.KcUser, error) {
	return restCall[[]*model.KcUser](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/users", realm), Query: restQuery("briefRepresentation", briefRepresentation, "email", email, "emailVerified", emailVerified, "enabled", enabled, "exact", exact, "first", first, "firstName", firstName, "idpAlias", idpAlias, "idpUserId", idpUserID, "lastName", lastName, "max", max, "q", q, "search", search, "username", username)}, "")
}

func (r *queryResolver) KcUserCount(ctx context.Context, realm string, briefRepresentation *string, email *string, emailVerified *string, enabled *string, exact *string, first *string, firstName *string, idpAlias *string, idpUserID *string, lastName *string, max *string, q *string, search *string, username *string) (*int, error) {
	return restCall[*int](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/users/count", realm), Query: restQuery("briefRepresentation", briefRepresentation, "email", email, "emailVerified", emailVerified, "enabled", enabled, "exact", exact, "first", first, "firstName", firstName, "idpAlias", idpAlias, "idpUserId", idpUserID, "lastName", lastName, "max", max, "q", q, "search", search, "username", username)}, "")
}

func (r *queryResolver) KcUserByID(ctx context.Context, realm string, userID string) (*model.KcUser, error) {
	return restCall[*model.KcUser](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/users/%v", realm, userID)}, "")
}

func (r *queryResolver) KcUserBruteForceDetectionStatus(ctx context.Context, realm string, userID string) (*model.KcBruteForceStatus, error) {
	return restCall[*model.KcBruteForceStatus](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/users/%v/brute-force", realm, userID)}, "")
}

func (r *queryResolver) KcUserSessions(ctx context.Context, realm string, userID string) ([]*model.KcUserSessionRepresentation, error) {
	return restCall[[]*model.KcUserSessionRepresentation](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/users/%v/sessions", realm, userID)}, "")
}

func (r *queryResolver) KcUserOfflineSessionsForClient(ctx context.Context, realm string, userID string, clientID string) ([]*model.KcUserSessionRepresentation, error) {
	return restCall[[]*model.KcUserSessionRepresentation](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/users/%v/offline-sessions/%v", realm, userID, clientID)}, "")
}

func (r *queryResolver) KcUserGroups(ctx context.Context, realm string, userID string, briefRepresentation *string, exact *string, first *string, full *string, max *string, q *string, search *string) ([]*model.KcGroup, error) {
	return restCall[[]*model.KcGroup](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/users/%v/groups", realm, userID), Query: restQuery("briefRepresentation", briefRepresentation, "exact", exact, "first", first, "full", full, "max", max, "q", q, "search", search)}, "")
}

func (r *queryResolver) KcCredentials(ctx context.Context, realm string, userID string) ([]*model.KcCredentialRepresentation, error) {
	return restCall[[]*model.KcCredentialRepresentation](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/users/%v/credentials", realm, userID)}, "")
}

func (r *queryResolver) KcConfiguredUserStorageCredentialTypes(ctx context.Context, realm string, userID string) ([]string, error) {
	return restCall[[]string](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/users/%v/configured-user-storage-credential-types", realm, userID)}, "")
}

func (r *queryResolver) KcUserFederatedIdentities(ctx context.Context, realm string, userID string) ([]*model.KcFederatedIdentityRepresentation, error) {
	return restCall[[]*model.KcFederatedIdentityRepresentation](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/users/%v/federated-identity", realm, userID)}, "")
}

func (r *queryResolver) KcGroups(ctx context.Context, realm string, briefRepresentation *string, exact *string, first *string, full *string, max *string, q *string, search *string) ([]*model.KcGroup, error) {
	return restCall[[]*model.KcGroup](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/groups", realm), Query: restQuery("briefRepresentation", briefRepresentation, "exact", exact, "first", first, "full", full, "max", max, "q", q, "search", search)}, "")
}

func (r *queryResolver) KcGroupsCount(ctx context.Context, realm string, briefRepresentation *string, exact *string, first *string, full *string, max *string, q *string, search *string) (*int, error) {
	return restCall[*int](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/groups/count", realm), Query: restQuery("briefRepresentation", briefRepresentation, "exact", exact, "first", first, "full", full, "max", max, "q", q, "search", search)}, "")
}

func (r *queryResolver) KcGroup(ctx context.Context, realm string, groupID string) (*model.KcGroup, error) {
	return restCall[*model.KcGroup](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/groups/%v", realm, groupID)}, "")
}

func (r *queryResolver) KcGroupMembers(ctx context.Context, realm string, groupID string, briefRepresentation *string, exact *string, first *string, full *string, max *string, q *string, search *string) ([]*model.KcUser, error) {
	return restCall[[]*model.KcUser](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/groups/%v/members", realm, groupID), Query: restQuery("briefRepresentation", briefRepresentation, "exact", exact, "first", first, "full", full, "max", max, "q", q, "search", search)}, "")
}

func (r *queryResolver) KcGroupByPath(ctx context.Context, realm string, path string) (*model.KcGroup, error) {
	return restCall[*model.KcGroup](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/group-by-path/%s", realm, rest.Raw(path))}, "")
}

func (r *queryResolver) KcDefaultGroups(ctx context.Context, realm string) ([]*model.KcGroup, error) {
	return restCall[[]*model.KcGroup](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/default-groups", realm)}, "")
}

func (r *queryResolver) KcRealmRoles(ctx context.Context, realm string, first *string, max *string, search *string, briefRepresentation *string) ([]*model.KcRole, error) {
	return restCall[[]*model.KcRole](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/roles", realm), Query: restQuery("first", first, "max", max, "search", search, "briefRepresentation", briefRepresentation)}, "")
}

func (r *queryResolver) KcRealmRole(ctx context.Context, realm string, roleName string) (*model.KcRole, error) {
	return restCall[*model.KcRole](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/roles/%v", realm, roleName)}, "")
}

func (r *queryResolver) KcCompositeRealmRoles(ctx context.Context, realm string, roleName string) ([]*model.KcRole, error) {
	return restCall[[]*model.KcRole](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/roles/%v/composites", realm, roleName)}, "")
}

func (r *queryResolver) KcUsersByRoleName(ctx context.Context, realm string, roleName string, first *string, max *string) ([]*model.KcUser, error) {
	return restCall[[]*model.KcUser](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/roles/%v/users", realm, roleName), Query: restQuery("first", first, "max", max)}, "")
}

func (r *queryResolver) KcGroupsByRole(ctx context.Context, realm string, roleName string) ([]*model.KcGroup, error) {
	return restCall[[]*model.KcGroup](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/roles/%v/groups", realm, roleName)}, "")
}

func (r *queryResolver) KcRealmRoleByID(ctx context.Context, realm string, roleID string) (*model.KcRole, error) {
	return restCall[*model.KcRole](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/roles-by-id/%v", realm, roleID)}, "")
}

func (r *queryResolver) KcCompositeRolesByRoleID(ctx context.Context, realm string, roleID string) ([]*model.KcRole, error) {
	return restCall[[]*model.KcRole](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/roles-by-id/%v/composites", realm, roleID)}, "")
}

func (r *queryResolver) KcCompositeRealmRolesByRoleID(ctx context.Context, realm string, roleID string) ([]*model.KcRole, error) {
	return restCall[[]*model.KcRole](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/roles-by-id/%v/composites/realm", realm, roleID)}, "")
}

func (r *queryResolver) KcCompositeClientRolesByRoleID(ctx context.Context, realm string, roleID string, clientID string) ([]*model.KcRole, error) {
	return restCall[[]*model.KcRole](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/roles-by-id/%v/composites/clients/%v", realm, roleID, clientID)}, "")
}

func (r *queryResolver) KcClientsWithParams(ctx context.Context, realm string, clientID *string, viewableOnly *string, first *string, max *string, search *string, q *string) ([]*model.KcClient, error) {
	return restCall[[]*model.KcClient](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/clients", realm), Query: restQuery("clientId", clientID, "viewableOnly", viewableOnly, "first", first, "max", max, "search", search, "q", q)}, "")
}

func (r *queryResolver) KcClient(ctx context.Context, realm string, clientID string) (*model.KcClient, error) {
	return restCall[*model.KcClient](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/clients/%v", realm, clientID)}, "")
}

func (r *queryResolver) KcClientSecret(ctx context.Context, realm string, clientID string) (*model.KcCredentialRepresentation, error) {
	return restCall[*model.KcCredentialRepresentation](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/clients/%v/client-secret", realm, clientID)}, "")
}

func (r *queryResolver) KcClientServiceAccount(ctx context.Context, realm string, clientID string) (*model.KcUser, error) {
	return restCall[*model.KcUser](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/clients/%v/service-account-user", realm, clientID)}, "")
}

func (r *queryResolver) KcClientUserSessions(ctx context.Context, realm string, clientID string) ([]*model.KcUserSessionRepresentation, error) {
	return restCall[[]*model.KcUserSessionRepresentation](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/clients/%v/user-sessions", realm, clientID)}, "")
}

func (r *queryResolver) KcClientOfflineSessions(ctx context.Context, realm string, clientID string) ([]*model.KcUserSessionRepresentation, error) {
	return restCall[[]*model.KcUserSessionRepresentation](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/clients/%v/offline-sessions", realm, clientID)}, "")
}

func (r *queryResolver) KcClientsDefaultScopes(ctx context.Context, realm string, clientID string) ([]*model.KcClientScope, error) {
	return restCall[[]*model.KcClientScope](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/clients/%v/default-client-scopes", realm, clientID)}, "")
}

func (r *queryResolver) KcClientsOptionalScopes(ctx context.Context, realm string, clientID string) ([]*model.KcClientScope, error) {
	return restCall[[]*model.KcClientScope](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/clients/%v/optional-client-scopes", realm, clientID)}, "")
}

func (r *queryResolver) KcClientScopeMappings(ctx context.Context, realm string, clientID string) (*model.KcMappingsRepresentation, error) {
	return restCall[*model.KcMappingsRepresentation](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/clients/%v/scope-mappings", realm, clientID)}, "")
}

func (r *queryResolver) KcClientRoles(ctx context.Context, realm string, clientID string, first *string, max *string, search *string, briefRepresentation *string) ([]*model.KcRole, error) {
	return restCall[[]*model.KcRole](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/clients/%v/roles", realm, clientID), Query: restQuery("first", first, "max", max, "search", search, "briefRepresentation", briefRepresentation)}, "")
}

func (r *queryResolver) KcClientRole(ctx context.Context, realm string, clientID string, roleName string) (*model.KcRole, error) {
	return restCall[*model.KcRole](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/clients/%v/roles/%v", realm, clientID, roleName)}, "")
}

func (r *queryResolver) KcUsersByClientRoleName(ctx context.Context, realm string, clientID string, roleName string, first *string, max *string) ([]*model.KcUser, error) {
	return restCall[[]*model.KcUser](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/clients/%v/roles/%v/users", realm, clientID, roleName), Query: restQuery("first", first, "max", max)}, "")
}

func (r *queryResolver) KcGroupsByClientRole(ctx context.Context, realm string, clientID string, roleName string) ([]*model.KcGroup, error) {
	return restCall[[]*model.KcGroup](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/clients/%v/roles/%v/groups", realm, clientID, roleName)}, "")
}

func (r *queryResolver) KcResourceServer(ctx context.Context, realm string, clientID string) (*model.KcResourceServerRepresentation, error) {
	return restCall[*model.KcResourceServerRepresentation](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/clients/%v/authz/resource-server", realm, clientID)}, "")
}

func (r *queryResolver) KcResources(ctx context.Context, realm string, clientID string, deep *string, first *string, max *string, name *string, owner *string, typeArg *string, uri *string, scope *string, matchingURI *string, exactName *string) ([]*model.KcResourceRepresentation, error) {
	return restCall[[]*model.KcResourceRepresentation](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/clients/%v/authz/resource-server/resource", realm, clientID), Query: restQuery("deep", deep, "first", first, "max", max, "name", name, "owner", owner, "type", typeArg, "uri", uri, "scope", scope, "matchingUri", matchingURI, "exactName", exactName)}, "")
}

func (r *queryResolver) KcResource(ctx context.Context, realm string, clientID string, resourceID string) (*model.KcResourceRepresentation, error) {
	return restCall[*model.KcResourceRepresentation](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/clients/%v/authz/resource-server/resource/%v", realm, clientID, resourceID)}, "")
}

func (r *queryResolver) KcScopes(ctx context.Context, realm string, clientID string, deep *string, first *string, max *string, name *string) ([]*model.KcScopeRepresentation, error) {
	return restCall[[]*model.KcScopeRepresentation](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/clients/%v/authz/resource-server/scope", realm, clientID), Query: restQuery("deep", deep, "first", first, "max", max, "name", name)}, "")
}

func (r *queryResolver) KcScope(ctx context.Context, realm string, clientID string, scopeID string) (*model.KcScopeRepresentation, error) {
	return restCall[*model.KcScopeRepresentation](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/clients/%v/authz/resource-server/scope/%v", realm, clientID, scopeID)}, "")
}

func (r *queryResolver) KcPolicies(ctx context.Context, realm string, clientID string, first *string, max *string, name *string, permission *string, typeArg *string) ([]*model.KcPolicyRepresentation, error) {
	return restCall[[]*model.KcPolicyRepresentation](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/clients/%v/authz/resource-server/policy", realm, clientID), Query: restQuery("first", first, "max", max, "name", name, "permission", permission, "type", typeArg)}, "")
}

func (r *queryResolver) KcPolicy(ctx context.Context, realm string, clientID string, policyID string) (*model.KcPolicyRepresentation, error) {
	return restCall[*model.KcPolicyRepresentation](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/clients/%v/authz/resource-server/policy/%v", realm, clientID, policyID)}, "")
}

func (r *queryResolver) KcAuthorizationPolicyAssociatedPolicies(ctx context.Context, realm string, clientID string, policyID string) ([]*model.KcPolicyRepresentation, error) {
	return restCall[[]*model.KcPolicyRepresentation](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/clients/%v/authz/resource-server/policy/%v/associatedPolicies", realm, clientID, policyID)}, "")
}

func (r *queryResolver) KcAuthorizationPolicyResources(ctx context.Context, realm string, clientID string, policyID string) ([]*model.KcPolicyResourceRepresentation, error) {
	return restCall[[]*model.KcPolicyResourceRepresentation](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/clients/%v/authz/resource-server/policy/%v/resources", realm, clientID, policyID)}, "")
}

func (r *queryResolver) KcAuthorizationPolicyScopes(ctx context.Context, realm string, clientID string, policyID string) ([]*model.KcPolicyScopeRepresentation, error) {
	return restCall[[]*model.KcPolicyScopeRepresentation](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/clients/%v/authz/resource-server/policy/%v/scopes", realm, clientID, policyID)}, "")
}

func (r *queryResolver) KcDependentPermissions(ctx context.Context, realm string, clientID string, policyID string) ([]*model.KcPermissionRepresentation, error) {
	return restCall[[]*model.KcPermissionRepresentation](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/clients/%v/authz/resource-server/policy/%v/dependentPolicies", realm, clientID, policyID)}, "")
}

func (r *queryResolver) KcPermissions(ctx context.Context, realm string, clientID string, first *string, max *string, name *string, resource *string, scope *string, typeArg *string) ([]*model.KcPermissionRepresentation, error) {
	return restCall[[]*model.KcPermissionRepresentation](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/clients/%v/authz/resource-server/permission", realm, clientID), Query: restQuery("first", first, "max", max, "name", name, "resource", resource, "scope", scope, "type", typeArg)}, "")
}

func (r *queryResolver) KcPermission(ctx context.Context, realm string, clientID string, permissionID string) (*model.KcPermissionRepresentation, error) {
	return restCall[*model.KcPermissionRepresentation](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/clients/%v/authz/resource-server/permission/%v", realm, clientID, permissionID)}, "")
}

func (r *queryResolver) KcPermissionResources(ctx context.Context, realm string, clientID string, permissionID string) ([]*model.KcPermissionResource, error) {
	return restCall[[]*model.KcPermissionResource](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/clients/%v/authz/resource-server/permission/%v/resources", realm, clientID, permissionID)}, "")
}

func (r *queryResolver) KcPermissionScopes(ctx context.Context, realm string, clientID string, permissionID string) ([]*model.KcPermissionScope, error) {
	return restCall[[]*model.KcPermissionScope](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/clients/%v/authz/resource-server/permission/%v/scopes", realm, clientID, permissionID)}, "")
}

func (r *queryResolver) KcClientRepresentation(ctx context.Context, realm string, clientID string) (*model.KcClient, error) {
	return restCall[*model.KcClient](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/clients-registrations/%v", realm, clientID)}, "")
}

func (r *queryResolver) KcAdapterConfiguration(ctx context.Context, realm string, clientID string) (*model.KcAdapterConfiguration, error) {
	return restCall[*model.KcAdapterConfiguration](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/clients-registrations/%v/installation", realm, clientID)}, "")
}

func (r *queryResolver) KcClientScopes(ctx context.Context, realm string) ([]*model.KcClientScope, error) {
	return restCall[[]*model.KcClientScope](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/client-scopes", realm)}, "")
}

func (r *queryResolver) KcClientScope(ctx context.Context, realm string, scopeID string) (*model.KcClientScope, error) {
	return restCall[*model.KcClientScope](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/client-scopes/%v", realm, scopeID)}, "")
}

func (r *queryResolver) KcClientScopeProtocolMappers(ctx context.Context, realm string, scopeID string) ([]*model.KcProtocolMappers, error) {
	return restCall[[]*model.KcProtocolMappers](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/client-scopes/%v/protocol-mappers", realm, scopeID)}, "")
}

func (r *queryResolver) KcClientScopeProtocolMapper(ctx context.Context, realm string, scopeID string, mapperID string) (*model.KcProtocolMappers, error) {
	return restCall[*model.KcProtocolMappers](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/client-scopes/%v/protocol-mappers/%v", realm, scopeID, mapperID)}, "")
}

func (r *queryResolver) KcDefaultDefaultClientScopes(ctx context.Context, realm string) ([]*model.KcClientScope, error) {
	return restCall[[]*model.KcClientScope](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/default-default-client-scopes", realm)}, "")
}

func (r *queryResolver) KcDefaultOptionalClientScopes(ctx context.Context, realm string) ([]*model.KcClientScope, error) {
	return restCall[[]*model.KcClientScope](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/default-optional-client-scopes", realm)}, "")
}

func (r *queryResolver) KcComponentsWithParams(ctx context.Context, realm string, name *string, provider *string, parent *string) ([]*model.KcComponent, error) {
	return restCall[[]*model.KcComponent](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/components", realm), Query: restQuery("name", name, "provider", provider, "parent", parent)}, "")
}

func (r *queryResolver) KcComponent(ctx context.Context, realm string, componentID string) (*model.KcComponent, error) {
	return restCall[*model.KcComponent](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/components/%v", realm, componentID)}, "")
}

func (r *queryResolver) KcAuthenticationFlows(ctx context.Context, realm string) ([]*model.KcAuthenticationFlowRepresentation, error) {
	return restCall[[]*model.KcAuthenticationFlowRepresentation](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/authentication/flows", realm)}, "")
}

func (r *queryResolver) KcAuthenticationFlow(ctx context.Context, realm string, flow string) (*model.KcAuthenticationFlowRepresentation, error) {
	return restCall[*model.KcAuthenticationFlowRepresentation](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/authentication/flows/%v", realm, flow)}, "")
}

func (r *queryResolver) KcAuthenticationExecutions(ctx context.Context, realm string, flow string) ([]*model.KcModifyAuthenticationExecutionRepresentation, error) {
	return restCall[[]*model.KcModifyAuthenticationExecutionRepresentation](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/authentication/flows/%v/executions", realm, flow)}, "")
}

func (r *queryResolver) KcRequiredActions(ctx context.Context, realm string) ([]*model.KcRequiredActionProviderRepresentation, error) {
	return restCall[[]*model.KcRequiredActionProviderRepresentation](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/authentication/required-actions", realm)}, "")
}

func (r *queryResolver) KcRequiredAction(ctx context.Context, realm string, alias string) (*model.KcRequiredActionProviderRepresentation, error) {
	return restCall[*model.KcRequiredActionProviderRepresentation](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/authentication/required-actions/%v", realm, alias)}, "")
}

func (r *queryResolver) KcIdentityProviders(ctx context.Context, realm string) ([]*model.KcIdentityProviderRepresentation, error) {
	return restCall[[]*model.KcIdentityProviderRepresentation](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/identity-providers", realm)}, "")
}

func (r *queryResolver) KcIdentityProvider(ctx context.Context, realm string, alias string) (*model.KcIdentityProviderRepresentation, error) {
	return restCall[*model.KcIdentityProviderRepresentation](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/identity-providers/%v", realm, alias)}, "")
}

func (r *queryResolver) KcExportIDPPublicBrokerConfig(ctx context.Context, realm string, alias string) (*string, error) {
	return restCall[*string](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/identity-providers/%v/export", realm, alias)}, "")
}

func (r *queryResolver) KcIdentityProviderMappers(ctx context.Context, realm string, alias string) ([]*model.KcIdentityProviderMapper, error) {
	return restCall[[]*model.KcIdentityProviderMapper](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/identity-providers/%v/mappers", realm, alias)}, "")
}

func (r *queryResolver) KcIdentityProviderMapperByID(ctx context.Context, realm string, alias string, mapperID string) (*model.KcIdentityProviderMapper, error) {
	return restCall[*model.KcIdentityProviderMapper](ctx, r.Proxy, "manage-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/admin/realms/%v/identity-providers/%v/mappers/%v", realm, alias, mapperID)}, "")
}
