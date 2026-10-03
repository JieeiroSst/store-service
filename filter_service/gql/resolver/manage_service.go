package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) KcQuery() generated.KcQueryResolver { return &kcQueryResolver{r} }

type kcQueryResolver struct{ *Resolver }

func (r *kcQueryResolver) TokenUser(ctx context.Context, obj *model.KcQuery, realm *string) (*model.KcTokenInfo, error) {
	return r.Clients.ManageService.TokenUser(ctx, realm)
}

func (r *kcQueryResolver) Issuer(ctx context.Context, obj *model.KcQuery, realm string) (*model.KcIssuerResponse, error) {
	return r.Clients.ManageService.Issuer(ctx, realm)
}

func (r *kcQueryResolver) Certs(ctx context.Context, obj *model.KcQuery, realm string) (*model.KcCertResponse, error) {
	return r.Clients.ManageService.Certs(ctx, realm)
}

func (r *kcQueryResolver) RawUserInfo(ctx context.Context, obj *model.KcQuery, realm string) (map[string]interface{}, error) {
	return r.Clients.ManageService.RawUserInfo(ctx, realm)
}

func (r *kcQueryResolver) DecodeAccessToken(ctx context.Context, obj *model.KcQuery, realm string) (map[string]interface{}, error) {
	return r.Clients.ManageService.DecodeAccessToken(ctx, realm)
}

func (r *kcQueryResolver) UserPermissions(ctx context.Context, obj *model.KcQuery, realm string) ([]*model.KcPermissionGrantResponseRepresentation, error) {
	return r.Clients.ManageService.UserPermissions(ctx, realm)
}

func (r *kcQueryResolver) ResourcesClient(ctx context.Context, obj *model.KcQuery, realm string, deep *string, first *string, max *string, name *string, owner *string, typeArg *string, uri *string, scope *string, matchingURI *string, exactName *string) ([]*model.KcResourceRepresentation, error) {
	return r.Clients.ManageService.ResourcesClient(ctx, realm, deep, first, max, name, owner, typeArg, uri, scope, matchingURI, exactName)
}

func (r *kcQueryResolver) ResourceClient(ctx context.Context, obj *model.KcQuery, realm string, resourceID string) (*model.KcResourceRepresentation, error) {
	return r.Clients.ManageService.ResourceClient(ctx, realm, resourceID)
}

func (r *kcQueryResolver) ResourcePolicies(ctx context.Context, obj *model.KcQuery, realm string, resource *string, name *string, scope *string, first *string, max *string) ([]*model.KcResourcePolicyRepresentation, error) {
	return r.Clients.ManageService.ResourcePolicies(ctx, realm, resource, name, scope, first, max)
}

func (r *kcQueryResolver) ResourcePolicy(ctx context.Context, obj *model.KcQuery, realm string, permissionID string) (*model.KcResourcePolicyRepresentation, error) {
	return r.Clients.ManageService.ResourcePolicy(ctx, realm, permissionID)
}

func (r *kcQueryResolver) ServerInfo(ctx context.Context, obj *model.KcQuery) (*model.KcServerInfoRepresentation, error) {
	return r.Clients.ManageService.ServerInfo(ctx)
}

func (r *kcQueryResolver) Realms(ctx context.Context, obj *model.KcQuery) ([]*model.KcRealmRepresentation, error) {
	return r.Clients.ManageService.Realms(ctx)
}

func (r *kcQueryResolver) Realm(ctx context.Context, obj *model.KcQuery, realm string) (*model.KcRealmRepresentation, error) {
	return r.Clients.ManageService.Realm(ctx, realm)
}

func (r *kcQueryResolver) KeyStoreConfig(ctx context.Context, obj *model.KcQuery, realm string) (*model.KcKeyStoreConfig, error) {
	return r.Clients.ManageService.KeyStoreConfig(ctx, realm)
}

func (r *kcQueryResolver) Events(ctx context.Context, obj *model.KcQuery, realm string, client *string, dateFrom *string, dateTo *string, first *string, ipAddress *string, max *string, user *string) ([]*model.KcEventRepresentation, error) {
	return r.Clients.ManageService.Events(ctx, realm, client, dateFrom, dateTo, first, ipAddress, max, user)
}

func (r *kcQueryResolver) CredentialRegistrators(ctx context.Context, obj *model.KcQuery, realm string) ([]string, error) {
	return r.Clients.ManageService.CredentialRegistrators(ctx, realm)
}

func (r *kcQueryResolver) Users(ctx context.Context, obj *model.KcQuery, realm string, briefRepresentation *string, email *string, emailVerified *string, enabled *string, exact *string, first *string, firstName *string, idpAlias *string, idpUserID *string, lastName *string, max *string, qArg *string, search *string, username *string) ([]*model.KcUser, error) {
	return r.Clients.ManageService.Users(ctx, realm, briefRepresentation, email, emailVerified, enabled, exact, first, firstName, idpAlias, idpUserID, lastName, max, qArg, search, username)
}

func (r *kcQueryResolver) UserCount(ctx context.Context, obj *model.KcQuery, realm string, briefRepresentation *string, email *string, emailVerified *string, enabled *string, exact *string, first *string, firstName *string, idpAlias *string, idpUserID *string, lastName *string, max *string, qArg *string, search *string, username *string) (*int, error) {
	return r.Clients.ManageService.UserCount(ctx, realm, briefRepresentation, email, emailVerified, enabled, exact, first, firstName, idpAlias, idpUserID, lastName, max, qArg, search, username)
}

func (r *kcQueryResolver) UserByID(ctx context.Context, obj *model.KcQuery, realm string, userID string) (*model.KcUser, error) {
	return r.Clients.ManageService.UserByID(ctx, realm, userID)
}

func (r *kcQueryResolver) UserBruteForceDetectionStatus(ctx context.Context, obj *model.KcQuery, realm string, userID string) (*model.KcBruteForceStatus, error) {
	return r.Clients.ManageService.UserBruteForceDetectionStatus(ctx, realm, userID)
}

func (r *kcQueryResolver) UserSessions(ctx context.Context, obj *model.KcQuery, realm string, userID string) ([]*model.KcUserSessionRepresentation, error) {
	return r.Clients.ManageService.UserSessions(ctx, realm, userID)
}

func (r *kcQueryResolver) UserOfflineSessionsForClient(ctx context.Context, obj *model.KcQuery, realm string, userID string, clientID string) ([]*model.KcUserSessionRepresentation, error) {
	return r.Clients.ManageService.UserOfflineSessionsForClient(ctx, realm, userID, clientID)
}

func (r *kcQueryResolver) UserGroups(ctx context.Context, obj *model.KcQuery, realm string, userID string, briefRepresentation *string, exact *string, first *string, full *string, max *string, qArg *string, search *string) ([]*model.KcGroup, error) {
	return r.Clients.ManageService.UserGroups(ctx, realm, userID, briefRepresentation, exact, first, full, max, qArg, search)
}

func (r *kcQueryResolver) Credentials(ctx context.Context, obj *model.KcQuery, realm string, userID string) ([]*model.KcCredentialRepresentation, error) {
	return r.Clients.ManageService.Credentials(ctx, realm, userID)
}

func (r *kcQueryResolver) ConfiguredUserStorageCredentialTypes(ctx context.Context, obj *model.KcQuery, realm string, userID string) ([]string, error) {
	return r.Clients.ManageService.ConfiguredUserStorageCredentialTypes(ctx, realm, userID)
}

func (r *kcQueryResolver) UserFederatedIdentities(ctx context.Context, obj *model.KcQuery, realm string, userID string) ([]*model.KcFederatedIdentityRepresentation, error) {
	return r.Clients.ManageService.UserFederatedIdentities(ctx, realm, userID)
}

func (r *kcQueryResolver) Groups(ctx context.Context, obj *model.KcQuery, realm string, briefRepresentation *string, exact *string, first *string, full *string, max *string, qArg *string, search *string) ([]*model.KcGroup, error) {
	return r.Clients.ManageService.Groups(ctx, realm, briefRepresentation, exact, first, full, max, qArg, search)
}

func (r *kcQueryResolver) GroupsCount(ctx context.Context, obj *model.KcQuery, realm string, briefRepresentation *string, exact *string, first *string, full *string, max *string, qArg *string, search *string) (*int, error) {
	return r.Clients.ManageService.GroupsCount(ctx, realm, briefRepresentation, exact, first, full, max, qArg, search)
}

func (r *kcQueryResolver) Group(ctx context.Context, obj *model.KcQuery, realm string, groupID string) (*model.KcGroup, error) {
	return r.Clients.ManageService.Group(ctx, realm, groupID)
}

func (r *kcQueryResolver) GroupMembers(ctx context.Context, obj *model.KcQuery, realm string, groupID string, briefRepresentation *string, exact *string, first *string, full *string, max *string, qArg *string, search *string) ([]*model.KcUser, error) {
	return r.Clients.ManageService.GroupMembers(ctx, realm, groupID, briefRepresentation, exact, first, full, max, qArg, search)
}

func (r *kcQueryResolver) GroupByPath(ctx context.Context, obj *model.KcQuery, realm string, pathArg string) (*model.KcGroup, error) {
	return r.Clients.ManageService.GroupByPath(ctx, realm, pathArg)
}

func (r *kcQueryResolver) DefaultGroups(ctx context.Context, obj *model.KcQuery, realm string) ([]*model.KcGroup, error) {
	return r.Clients.ManageService.DefaultGroups(ctx, realm)
}

func (r *kcQueryResolver) RealmRoles(ctx context.Context, obj *model.KcQuery, realm string, first *string, max *string, search *string, briefRepresentation *string) ([]*model.KcRole, error) {
	return r.Clients.ManageService.RealmRoles(ctx, realm, first, max, search, briefRepresentation)
}

func (r *kcQueryResolver) RealmRole(ctx context.Context, obj *model.KcQuery, realm string, roleName string) (*model.KcRole, error) {
	return r.Clients.ManageService.RealmRole(ctx, realm, roleName)
}

func (r *kcQueryResolver) CompositeRealmRoles(ctx context.Context, obj *model.KcQuery, realm string, roleName string) ([]*model.KcRole, error) {
	return r.Clients.ManageService.CompositeRealmRoles(ctx, realm, roleName)
}

func (r *kcQueryResolver) UsersByRoleName(ctx context.Context, obj *model.KcQuery, realm string, roleName string, first *string, max *string) ([]*model.KcUser, error) {
	return r.Clients.ManageService.UsersByRoleName(ctx, realm, roleName, first, max)
}

func (r *kcQueryResolver) GroupsByRole(ctx context.Context, obj *model.KcQuery, realm string, roleName string) ([]*model.KcGroup, error) {
	return r.Clients.ManageService.GroupsByRole(ctx, realm, roleName)
}

func (r *kcQueryResolver) RealmRoleByID(ctx context.Context, obj *model.KcQuery, realm string, roleID string) (*model.KcRole, error) {
	return r.Clients.ManageService.RealmRoleByID(ctx, realm, roleID)
}

func (r *kcQueryResolver) CompositeRolesByRoleID(ctx context.Context, obj *model.KcQuery, realm string, roleID string) ([]*model.KcRole, error) {
	return r.Clients.ManageService.CompositeRolesByRoleID(ctx, realm, roleID)
}

func (r *kcQueryResolver) CompositeRealmRolesByRoleID(ctx context.Context, obj *model.KcQuery, realm string, roleID string) ([]*model.KcRole, error) {
	return r.Clients.ManageService.CompositeRealmRolesByRoleID(ctx, realm, roleID)
}

func (r *kcQueryResolver) CompositeClientRolesByRoleID(ctx context.Context, obj *model.KcQuery, realm string, roleID string, clientID string) ([]*model.KcRole, error) {
	return r.Clients.ManageService.CompositeClientRolesByRoleID(ctx, realm, roleID, clientID)
}

func (r *kcQueryResolver) ClientsWithParams(ctx context.Context, obj *model.KcQuery, realm string, clientID *string, viewableOnly *string, first *string, max *string, search *string, qArg *string) ([]*model.KcClient, error) {
	return r.Clients.ManageService.ClientsWithParams(ctx, realm, clientID, viewableOnly, first, max, search, qArg)
}

func (r *kcQueryResolver) Client(ctx context.Context, obj *model.KcQuery, realm string, clientID string) (*model.KcClient, error) {
	return r.Clients.ManageService.Client(ctx, realm, clientID)
}

func (r *kcQueryResolver) ClientSecret(ctx context.Context, obj *model.KcQuery, realm string, clientID string) (*model.KcCredentialRepresentation, error) {
	return r.Clients.ManageService.ClientSecret(ctx, realm, clientID)
}

func (r *kcQueryResolver) ClientServiceAccount(ctx context.Context, obj *model.KcQuery, realm string, clientID string) (*model.KcUser, error) {
	return r.Clients.ManageService.ClientServiceAccount(ctx, realm, clientID)
}

func (r *kcQueryResolver) ClientUserSessions(ctx context.Context, obj *model.KcQuery, realm string, clientID string) ([]*model.KcUserSessionRepresentation, error) {
	return r.Clients.ManageService.ClientUserSessions(ctx, realm, clientID)
}

func (r *kcQueryResolver) ClientOfflineSessions(ctx context.Context, obj *model.KcQuery, realm string, clientID string) ([]*model.KcUserSessionRepresentation, error) {
	return r.Clients.ManageService.ClientOfflineSessions(ctx, realm, clientID)
}

func (r *kcQueryResolver) ClientsDefaultScopes(ctx context.Context, obj *model.KcQuery, realm string, clientID string) ([]*model.KcClientScope, error) {
	return r.Clients.ManageService.ClientsDefaultScopes(ctx, realm, clientID)
}

func (r *kcQueryResolver) ClientsOptionalScopes(ctx context.Context, obj *model.KcQuery, realm string, clientID string) ([]*model.KcClientScope, error) {
	return r.Clients.ManageService.ClientsOptionalScopes(ctx, realm, clientID)
}

func (r *kcQueryResolver) ClientScopeMappings(ctx context.Context, obj *model.KcQuery, realm string, clientID string) (*model.KcMappingsRepresentation, error) {
	return r.Clients.ManageService.ClientScopeMappings(ctx, realm, clientID)
}

func (r *kcQueryResolver) ClientRoles(ctx context.Context, obj *model.KcQuery, realm string, clientID string, first *string, max *string, search *string, briefRepresentation *string) ([]*model.KcRole, error) {
	return r.Clients.ManageService.ClientRoles(ctx, realm, clientID, first, max, search, briefRepresentation)
}

func (r *kcQueryResolver) ClientRole(ctx context.Context, obj *model.KcQuery, realm string, clientID string, roleName string) (*model.KcRole, error) {
	return r.Clients.ManageService.ClientRole(ctx, realm, clientID, roleName)
}

func (r *kcQueryResolver) UsersByClientRoleName(ctx context.Context, obj *model.KcQuery, realm string, clientID string, roleName string, first *string, max *string) ([]*model.KcUser, error) {
	return r.Clients.ManageService.UsersByClientRoleName(ctx, realm, clientID, roleName, first, max)
}

func (r *kcQueryResolver) GroupsByClientRole(ctx context.Context, obj *model.KcQuery, realm string, clientID string, roleName string) ([]*model.KcGroup, error) {
	return r.Clients.ManageService.GroupsByClientRole(ctx, realm, clientID, roleName)
}

func (r *kcQueryResolver) ResourceServer(ctx context.Context, obj *model.KcQuery, realm string, clientID string) (*model.KcResourceServerRepresentation, error) {
	return r.Clients.ManageService.ResourceServer(ctx, realm, clientID)
}

func (r *kcQueryResolver) Resources(ctx context.Context, obj *model.KcQuery, realm string, clientID string, deep *string, first *string, max *string, name *string, owner *string, typeArg *string, uri *string, scope *string, matchingURI *string, exactName *string) ([]*model.KcResourceRepresentation, error) {
	return r.Clients.ManageService.Resources(ctx, realm, clientID, deep, first, max, name, owner, typeArg, uri, scope, matchingURI, exactName)
}

func (r *kcQueryResolver) Resource(ctx context.Context, obj *model.KcQuery, realm string, clientID string, resourceID string) (*model.KcResourceRepresentation, error) {
	return r.Clients.ManageService.Resource(ctx, realm, clientID, resourceID)
}

func (r *kcQueryResolver) Scopes(ctx context.Context, obj *model.KcQuery, realm string, clientID string, deep *string, first *string, max *string, name *string) ([]*model.KcScopeRepresentation, error) {
	return r.Clients.ManageService.Scopes(ctx, realm, clientID, deep, first, max, name)
}

func (r *kcQueryResolver) Scope(ctx context.Context, obj *model.KcQuery, realm string, clientID string, scopeID string) (*model.KcScopeRepresentation, error) {
	return r.Clients.ManageService.Scope(ctx, realm, clientID, scopeID)
}

func (r *kcQueryResolver) Policies(ctx context.Context, obj *model.KcQuery, realm string, clientID string, first *string, max *string, name *string, permission *string, typeArg *string) ([]*model.KcPolicyRepresentation, error) {
	return r.Clients.ManageService.Policies(ctx, realm, clientID, first, max, name, permission, typeArg)
}

func (r *kcQueryResolver) Policy(ctx context.Context, obj *model.KcQuery, realm string, clientID string, policyID string) (*model.KcPolicyRepresentation, error) {
	return r.Clients.ManageService.Policy(ctx, realm, clientID, policyID)
}

func (r *kcQueryResolver) AuthorizationPolicyAssociatedPolicies(ctx context.Context, obj *model.KcQuery, realm string, clientID string, policyID string) ([]*model.KcPolicyRepresentation, error) {
	return r.Clients.ManageService.AuthorizationPolicyAssociatedPolicies(ctx, realm, clientID, policyID)
}

func (r *kcQueryResolver) AuthorizationPolicyResources(ctx context.Context, obj *model.KcQuery, realm string, clientID string, policyID string) ([]*model.KcPolicyResourceRepresentation, error) {
	return r.Clients.ManageService.AuthorizationPolicyResources(ctx, realm, clientID, policyID)
}

func (r *kcQueryResolver) AuthorizationPolicyScopes(ctx context.Context, obj *model.KcQuery, realm string, clientID string, policyID string) ([]*model.KcPolicyScopeRepresentation, error) {
	return r.Clients.ManageService.AuthorizationPolicyScopes(ctx, realm, clientID, policyID)
}

func (r *kcQueryResolver) DependentPermissions(ctx context.Context, obj *model.KcQuery, realm string, clientID string, policyID string) ([]*model.KcPermissionRepresentation, error) {
	return r.Clients.ManageService.DependentPermissions(ctx, realm, clientID, policyID)
}

func (r *kcQueryResolver) Permissions(ctx context.Context, obj *model.KcQuery, realm string, clientID string, first *string, max *string, name *string, resource *string, scope *string, typeArg *string) ([]*model.KcPermissionRepresentation, error) {
	return r.Clients.ManageService.Permissions(ctx, realm, clientID, first, max, name, resource, scope, typeArg)
}

func (r *kcQueryResolver) Permission(ctx context.Context, obj *model.KcQuery, realm string, clientID string, permissionID string) (*model.KcPermissionRepresentation, error) {
	return r.Clients.ManageService.Permission(ctx, realm, clientID, permissionID)
}

func (r *kcQueryResolver) PermissionResources(ctx context.Context, obj *model.KcQuery, realm string, clientID string, permissionID string) ([]*model.KcPermissionResource, error) {
	return r.Clients.ManageService.PermissionResources(ctx, realm, clientID, permissionID)
}

func (r *kcQueryResolver) PermissionScopes(ctx context.Context, obj *model.KcQuery, realm string, clientID string, permissionID string) ([]*model.KcPermissionScope, error) {
	return r.Clients.ManageService.PermissionScopes(ctx, realm, clientID, permissionID)
}

func (r *kcQueryResolver) ClientRepresentation(ctx context.Context, obj *model.KcQuery, realm string, clientID string) (*model.KcClient, error) {
	return r.Clients.ManageService.ClientRepresentation(ctx, realm, clientID)
}

func (r *kcQueryResolver) AdapterConfiguration(ctx context.Context, obj *model.KcQuery, realm string, clientID string) (*model.KcAdapterConfiguration, error) {
	return r.Clients.ManageService.AdapterConfiguration(ctx, realm, clientID)
}

func (r *kcQueryResolver) ClientScopes(ctx context.Context, obj *model.KcQuery, realm string) ([]*model.KcClientScope, error) {
	return r.Clients.ManageService.ClientScopes(ctx, realm)
}

func (r *kcQueryResolver) ClientScope(ctx context.Context, obj *model.KcQuery, realm string, scopeID string) (*model.KcClientScope, error) {
	return r.Clients.ManageService.ClientScope(ctx, realm, scopeID)
}

func (r *kcQueryResolver) ClientScopeProtocolMappers(ctx context.Context, obj *model.KcQuery, realm string, scopeID string) ([]*model.KcProtocolMappers, error) {
	return r.Clients.ManageService.ClientScopeProtocolMappers(ctx, realm, scopeID)
}

func (r *kcQueryResolver) ClientScopeProtocolMapper(ctx context.Context, obj *model.KcQuery, realm string, scopeID string, mapperID string) (*model.KcProtocolMappers, error) {
	return r.Clients.ManageService.ClientScopeProtocolMapper(ctx, realm, scopeID, mapperID)
}

func (r *kcQueryResolver) DefaultDefaultClientScopes(ctx context.Context, obj *model.KcQuery, realm string) ([]*model.KcClientScope, error) {
	return r.Clients.ManageService.DefaultDefaultClientScopes(ctx, realm)
}

func (r *kcQueryResolver) DefaultOptionalClientScopes(ctx context.Context, obj *model.KcQuery, realm string) ([]*model.KcClientScope, error) {
	return r.Clients.ManageService.DefaultOptionalClientScopes(ctx, realm)
}

func (r *kcQueryResolver) ComponentsWithParams(ctx context.Context, obj *model.KcQuery, realm string, name *string, provider *string, parent *string) ([]*model.KcComponent, error) {
	return r.Clients.ManageService.ComponentsWithParams(ctx, realm, name, provider, parent)
}

func (r *kcQueryResolver) Component(ctx context.Context, obj *model.KcQuery, realm string, componentID string) (*model.KcComponent, error) {
	return r.Clients.ManageService.Component(ctx, realm, componentID)
}

func (r *kcQueryResolver) AuthenticationFlows(ctx context.Context, obj *model.KcQuery, realm string) ([]*model.KcAuthenticationFlowRepresentation, error) {
	return r.Clients.ManageService.AuthenticationFlows(ctx, realm)
}

func (r *kcQueryResolver) AuthenticationFlow(ctx context.Context, obj *model.KcQuery, realm string, flow string) (*model.KcAuthenticationFlowRepresentation, error) {
	return r.Clients.ManageService.AuthenticationFlow(ctx, realm, flow)
}

func (r *kcQueryResolver) AuthenticationExecutions(ctx context.Context, obj *model.KcQuery, realm string, flow string) ([]*model.KcModifyAuthenticationExecutionRepresentation, error) {
	return r.Clients.ManageService.AuthenticationExecutions(ctx, realm, flow)
}

func (r *kcQueryResolver) RequiredActions(ctx context.Context, obj *model.KcQuery, realm string) ([]*model.KcRequiredActionProviderRepresentation, error) {
	return r.Clients.ManageService.RequiredActions(ctx, realm)
}

func (r *kcQueryResolver) RequiredAction(ctx context.Context, obj *model.KcQuery, realm string, alias string) (*model.KcRequiredActionProviderRepresentation, error) {
	return r.Clients.ManageService.RequiredAction(ctx, realm, alias)
}

func (r *kcQueryResolver) IdentityProviders(ctx context.Context, obj *model.KcQuery, realm string) ([]*model.KcIdentityProviderRepresentation, error) {
	return r.Clients.ManageService.IdentityProviders(ctx, realm)
}

func (r *kcQueryResolver) IdentityProvider(ctx context.Context, obj *model.KcQuery, realm string, alias string) (*model.KcIdentityProviderRepresentation, error) {
	return r.Clients.ManageService.IdentityProvider(ctx, realm, alias)
}

func (r *kcQueryResolver) ExportIDPPublicBrokerConfig(ctx context.Context, obj *model.KcQuery, realm string, alias string) (*string, error) {
	return r.Clients.ManageService.ExportIDPPublicBrokerConfig(ctx, realm, alias)
}

func (r *kcQueryResolver) IdentityProviderMappers(ctx context.Context, obj *model.KcQuery, realm string, alias string) ([]*model.KcIdentityProviderMapper, error) {
	return r.Clients.ManageService.IdentityProviderMappers(ctx, realm, alias)
}

func (r *kcQueryResolver) IdentityProviderMapperByID(ctx context.Context, obj *model.KcQuery, realm string, alias string, mapperID string) (*model.KcIdentityProviderMapper, error) {
	return r.Clients.ManageService.IdentityProviderMapperByID(ctx, realm, alias, mapperID)
}
