package keycloak

import (
	"context"
	"io"

	"github.com/Nerzal/gocloak/v13"
)

func (r *IdentityProvider) GetToken(ctx context.Context, realm string, options gocloak.TokenOptions) (*gocloak.JWT, error) {
	return r.client.GetToken(ctx, realm, options)
}

func (r *IdentityProvider) RetrospectToken(ctx context.Context, accessToken, clientID, clientSecret, realm string) (*gocloak.IntroSpectTokenResult, error) {
	return r.client.RetrospectToken(ctx, accessToken, clientID, clientSecret, realm)
}

func (r *IdentityProvider) DecodeAccessToken(ctx context.Context, accessToken, realm string) (map[string]interface{}, error) {
	_, claims, err := r.client.DecodeAccessToken(ctx, accessToken, realm)
	if err != nil {
		return nil, err
	}
	return *claims, nil
}

func (r *IdentityProvider) RevokeToken(ctx context.Context, realm, clientID, clientSecret, refreshToken string) error {
	return r.client.RevokeToken(ctx, realm, clientID, clientSecret, refreshToken)
}

func (r *IdentityProvider) LogoutPublicClient(ctx context.Context, clientID, realm, accessToken, refreshToken string) error {
	return r.client.LogoutPublicClient(ctx, clientID, realm, accessToken, refreshToken)
}

func (r *IdentityProvider) LoginClientTokenExchange(ctx context.Context, clientID, token, clientSecret, realm, targetClient, userID string) (*gocloak.JWT, error) {
	return r.client.LoginClientTokenExchange(ctx, clientID, token, clientSecret, realm, targetClient, userID)
}

func (r *IdentityProvider) GetRawUserInfo(ctx context.Context, accessToken, realm string) (map[string]interface{}, error) {
	return r.client.GetRawUserInfo(ctx, accessToken, realm)
}

func (r *IdentityProvider) GetCerts(ctx context.Context, realm string) (*gocloak.CertResponse, error) {
	return r.client.GetCerts(ctx, realm)
}

func (r *IdentityProvider) GetIssuer(ctx context.Context, realm string) (*gocloak.IssuerResponse, error) {
	return r.client.GetIssuer(ctx, realm)
}

func (r *IdentityProvider) GetServerInfo(ctx context.Context, accessToken string) (*gocloak.ServerInfoRepresentation, error) {
	return r.client.GetServerInfo(ctx, accessToken)
}

func (r *IdentityProvider) CreateUserRepresentation(ctx context.Context, token, realm string, user gocloak.User) (string, error) {
	return r.client.CreateUser(ctx, token, realm, user)
}

func (r *IdentityProvider) ExecuteActionsEmail(ctx context.Context, token, realm string, params gocloak.ExecuteActionsEmail) error {
	return r.client.ExecuteActionsEmail(ctx, token, realm, params)
}

func (r *IdentityProvider) SendVerifyEmail(ctx context.Context, token, userID, realm string, params gocloak.SendVerificationMailParams) error {
	return r.client.SendVerifyEmail(ctx, token, userID, realm, params)
}

func (r *IdentityProvider) GetUserBruteForceDetectionStatus(ctx context.Context, accessToken, realm, userID string) (*gocloak.BruteForceStatus, error) {
	return r.client.GetUserBruteForceDetectionStatus(ctx, accessToken, realm, userID)
}

func (r *IdentityProvider) RevokeUserConsents(ctx context.Context, accessToken, realm, userID, clientID string) error {
	return r.client.RevokeUserConsents(ctx, accessToken, realm, userID, clientID)
}

func (r *IdentityProvider) LogoutAllSessions(ctx context.Context, accessToken, realm, userID string) error {
	return r.client.LogoutAllSessions(ctx, accessToken, realm, userID)
}

func (r *IdentityProvider) LogoutUserSession(ctx context.Context, accessToken, realm, session string) error {
	return r.client.LogoutUserSession(ctx, accessToken, realm, session)
}

func (r *IdentityProvider) CreateChildGroup(ctx context.Context, token, realm, groupID string, group gocloak.Group) (string, error) {
	return r.client.CreateChildGroup(ctx, token, realm, groupID, group)
}

func (r *IdentityProvider) GetGroupByPath(ctx context.Context, token, realm, groupPath string) (*gocloak.Group, error) {
	return r.client.GetGroupByPath(ctx, token, realm, groupPath)
}

func (r *IdentityProvider) CreateRealmRole(ctx context.Context, token, realm string, role gocloak.Role) (string, error) {
	return r.client.CreateRealmRole(ctx, token, realm, role)
}

func (r *IdentityProvider) GetRealmRole(ctx context.Context, token, realm, roleName string) (*gocloak.Role, error) {
	return r.client.GetRealmRole(ctx, token, realm, roleName)
}

func (r *IdentityProvider) GetRealmRoleByID(ctx context.Context, token, realm, roleID string) (*gocloak.Role, error) {
	return r.client.GetRealmRoleByID(ctx, token, realm, roleID)
}

func (r *IdentityProvider) GetRealmRoles(ctx context.Context, token, realm string, params gocloak.GetRoleParams) ([]*gocloak.Role, error) {
	return r.client.GetRealmRoles(ctx, token, realm, params)
}

func (r *IdentityProvider) UpdateRealmRole(ctx context.Context, token, realm, roleName string, role gocloak.Role) error {
	return r.client.UpdateRealmRole(ctx, token, realm, roleName, role)
}

func (r *IdentityProvider) UpdateRealmRoleByID(ctx context.Context, token, realm, roleID string, role gocloak.Role) error {
	return r.client.UpdateRealmRoleByID(ctx, token, realm, roleID, role)
}

func (r *IdentityProvider) DeleteRealmRole(ctx context.Context, token, realm, roleName string) error {
	return r.client.DeleteRealmRole(ctx, token, realm, roleName)
}

func (r *IdentityProvider) AddRealmRoleComposite(ctx context.Context, token, realm, roleName string, roles []gocloak.Role) error {
	return r.client.AddRealmRoleComposite(ctx, token, realm, roleName, roles)
}

func (r *IdentityProvider) DeleteRealmRoleComposite(ctx context.Context, token, realm, roleName string, roles []gocloak.Role) error {
	return r.client.DeleteRealmRoleComposite(ctx, token, realm, roleName, roles)
}

func (r *IdentityProvider) GetCompositeRealmRoles(ctx context.Context, token, realm, roleName string) ([]*gocloak.Role, error) {
	return r.client.GetCompositeRealmRoles(ctx, token, realm, roleName)
}

func (r *IdentityProvider) GetCompositeRealmRolesByRoleID(ctx context.Context, token, realm, roleID string) ([]*gocloak.Role, error) {
	return r.client.GetCompositeRealmRolesByRoleID(ctx, token, realm, roleID)
}

func (r *IdentityProvider) GetCompositeRolesByRoleID(ctx context.Context, token, realm, roleID string) ([]*gocloak.Role, error) {
	return r.client.GetCompositeRolesByRoleID(ctx, token, realm, roleID)
}

func (r *IdentityProvider) GetCompositeClientRolesByRoleID(ctx context.Context, token, realm, idOfClient, roleID string) ([]*gocloak.Role, error) {
	return r.client.GetCompositeClientRolesByRoleID(ctx, token, realm, idOfClient, roleID)
}

func (r *IdentityProvider) GetGroupsByRole(ctx context.Context, token, realm, roleName string) ([]*gocloak.Group, error) {
	return r.client.GetGroupsByRole(ctx, token, realm, roleName)
}

func (r *IdentityProvider) CreateClientRole(ctx context.Context, token, realm, idOfClient string, role gocloak.Role) (string, error) {
	return r.client.CreateClientRole(ctx, token, realm, idOfClient, role)
}

func (r *IdentityProvider) GetGroupsByClientRole(ctx context.Context, token, realm, roleName, idOfClient string) ([]*gocloak.Group, error) {
	return r.client.GetGroupsByClientRole(ctx, token, realm, roleName, idOfClient)
}

func (r *IdentityProvider) AddRealmRoleToUser(ctx context.Context, token, realm, userID string, roles []gocloak.Role) error {
	return r.client.AddRealmRoleToUser(ctx, token, realm, userID, roles)
}

func (r *IdentityProvider) DeleteRealmRoleFromUser(ctx context.Context, token, realm, userID string, roles []gocloak.Role) error {
	return r.client.DeleteRealmRoleFromUser(ctx, token, realm, userID, roles)
}

func (r *IdentityProvider) GetRealmRolesByUserID(ctx context.Context, token, realm, userID string) ([]*gocloak.Role, error) {
	return r.client.GetRealmRolesByUserID(ctx, token, realm, userID)
}

func (r *IdentityProvider) GetCompositeRealmRolesByUserID(ctx context.Context, token, realm, userID string) ([]*gocloak.Role, error) {
	return r.client.GetCompositeRealmRolesByUserID(ctx, token, realm, userID)
}

func (r *IdentityProvider) GetAvailableRealmRolesByUserID(ctx context.Context, token, realm, userID string) ([]*gocloak.Role, error) {
	return r.client.GetAvailableRealmRolesByUserID(ctx, token, realm, userID)
}

func (r *IdentityProvider) AddClientRolesToUser(ctx context.Context, token, realm, idOfClient, userID string, roles []gocloak.Role) error {
	return r.client.AddClientRolesToUser(ctx, token, realm, idOfClient, userID, roles)
}

func (r *IdentityProvider) DeleteClientRolesFromUser(ctx context.Context, token, realm, idOfClient, userID string, roles []gocloak.Role) error {
	return r.client.DeleteClientRolesFromUser(ctx, token, realm, idOfClient, userID, roles)
}

func (r *IdentityProvider) GetClientRolesByUserID(ctx context.Context, token, realm, idOfClient, userID string) ([]*gocloak.Role, error) {
	return r.client.GetClientRolesByUserID(ctx, token, realm, idOfClient, userID)
}

func (r *IdentityProvider) GetCompositeClientRolesByUserID(ctx context.Context, token, realm, idOfClient, userID string) ([]*gocloak.Role, error) {
	return r.client.GetCompositeClientRolesByUserID(ctx, token, realm, idOfClient, userID)
}

func (r *IdentityProvider) GetAvailableClientRolesByUserID(ctx context.Context, token, realm, idOfClient, userID string) ([]*gocloak.Role, error) {
	return r.client.GetAvailableClientRolesByUserID(ctx, token, realm, idOfClient, userID)
}

func (r *IdentityProvider) AddRealmRoleToGroup(ctx context.Context, token, realm, groupID string, roles []gocloak.Role) error {
	return r.client.AddRealmRoleToGroup(ctx, token, realm, groupID, roles)
}

func (r *IdentityProvider) DeleteRealmRoleFromGroup(ctx context.Context, token, realm, groupID string, roles []gocloak.Role) error {
	return r.client.DeleteRealmRoleFromGroup(ctx, token, realm, groupID, roles)
}

func (r *IdentityProvider) GetRealmRolesByGroupID(ctx context.Context, token, realm, groupID string) ([]*gocloak.Role, error) {
	return r.client.GetRealmRolesByGroupID(ctx, token, realm, groupID)
}

func (r *IdentityProvider) GetCompositeRealmRolesByGroupID(ctx context.Context, token, realm, groupID string) ([]*gocloak.Role, error) {
	return r.client.GetCompositeRealmRolesByGroupID(ctx, token, realm, groupID)
}

func (r *IdentityProvider) GetAvailableRealmRolesByGroupID(ctx context.Context, token, realm, groupID string) ([]*gocloak.Role, error) {
	return r.client.GetAvailableRealmRolesByGroupID(ctx, token, realm, groupID)
}

func (r *IdentityProvider) AddClientRolesToGroup(ctx context.Context, token, realm, idOfClient, groupID string, roles []gocloak.Role) error {
	return r.client.AddClientRolesToGroup(ctx, token, realm, idOfClient, groupID, roles)
}

func (r *IdentityProvider) DeleteClientRoleFromGroup(ctx context.Context, token, realm, idOfClient, groupID string, roles []gocloak.Role) error {
	return r.client.DeleteClientRoleFromGroup(ctx, token, realm, idOfClient, groupID, roles)
}

func (r *IdentityProvider) GetClientRolesByGroupID(ctx context.Context, token, realm, idOfClient, groupID string) ([]*gocloak.Role, error) {
	return r.client.GetClientRolesByGroupID(ctx, token, realm, idOfClient, groupID)
}

func (r *IdentityProvider) GetCompositeClientRolesByGroupID(ctx context.Context, token, realm, idOfClient, groupID string) ([]*gocloak.Role, error) {
	return r.client.GetCompositeClientRolesByGroupID(ctx, token, realm, idOfClient, groupID)
}

func (r *IdentityProvider) GetAvailableClientRolesByGroupID(ctx context.Context, token, realm, idOfClient, groupID string) ([]*gocloak.Role, error) {
	return r.client.GetAvailableClientRolesByGroupID(ctx, token, realm, idOfClient, groupID)
}

func (r *IdentityProvider) GetClientsWithParams(ctx context.Context, token, realm string, params gocloak.GetClientsParams) ([]*gocloak.Client, error) {
	return r.client.GetClients(ctx, token, realm, params)
}

func (r *IdentityProvider) CreateClient(ctx context.Context, accessToken, realm string, newClient gocloak.Client) (string, error) {
	return r.client.CreateClient(ctx, accessToken, realm, newClient)
}

func (r *IdentityProvider) CreateClientRepresentation(ctx context.Context, token, realm string, newClient gocloak.Client) (*gocloak.Client, error) {
	return r.client.CreateClientRepresentation(ctx, token, realm, newClient)
}

func (r *IdentityProvider) GetClientRepresentation(ctx context.Context, accessToken, realm, clientID string) (*gocloak.Client, error) {
	return r.client.GetClientRepresentation(ctx, accessToken, realm, clientID)
}

func (r *IdentityProvider) UpdateClientRepresentation(ctx context.Context, accessToken, realm string, updatedClient gocloak.Client) (*gocloak.Client, error) {
	return r.client.UpdateClientRepresentation(ctx, accessToken, realm, updatedClient)
}

func (r *IdentityProvider) DeleteClientRepresentation(ctx context.Context, accessToken, realm, clientID string) error {
	return r.client.DeleteClientRepresentation(ctx, accessToken, realm, clientID)
}

func (r *IdentityProvider) GetAdapterConfiguration(ctx context.Context, accessToken, realm, clientID string) (*gocloak.AdapterConfiguration, error) {
	return r.client.GetAdapterConfiguration(ctx, accessToken, realm, clientID)
}

func (r *IdentityProvider) GetResourceServer(ctx context.Context, token, realm, idOfClient string) (*gocloak.ResourceServerRepresentation, error) {
	return r.client.GetResourceServer(ctx, token, realm, idOfClient)
}

func (r *IdentityProvider) CreateClientScope(ctx context.Context, token, realm string, scope gocloak.ClientScope) (string, error) {
	return r.client.CreateClientScope(ctx, token, realm, scope)
}

func (r *IdentityProvider) CreateClientScopeProtocolMapper(ctx context.Context, token, realm, scopeID string, protocolMapper gocloak.ProtocolMappers) (string, error) {
	return r.client.CreateClientScopeProtocolMapper(ctx, token, realm, scopeID, protocolMapper)
}

func (r *IdentityProvider) GetClientScopeProtocolMapper(ctx context.Context, token, realm, scopeID, protocolMapperID string) (*gocloak.ProtocolMappers, error) {
	return r.client.GetClientScopeProtocolMapper(ctx, token, realm, scopeID, protocolMapperID)
}

func (r *IdentityProvider) GetClientScopeProtocolMappers(ctx context.Context, token, realm, scopeID string) ([]*gocloak.ProtocolMappers, error) {
	return r.client.GetClientScopeProtocolMappers(ctx, token, realm, scopeID)
}

func (r *IdentityProvider) UpdateClientScopeProtocolMapper(ctx context.Context, token, realm, scopeID string, protocolMapper gocloak.ProtocolMappers) error {
	return r.client.UpdateClientScopeProtocolMapper(ctx, token, realm, scopeID, protocolMapper)
}

func (r *IdentityProvider) DeleteClientScopeProtocolMapper(ctx context.Context, token, realm, scopeID, protocolMapperID string) error {
	return r.client.DeleteClientScopeProtocolMapper(ctx, token, realm, scopeID, protocolMapperID)
}

func (r *IdentityProvider) CreateClientScopeMappingsRealmRoles(ctx context.Context, token, realm, idOfClient string, roles []gocloak.Role) error {
	return r.client.CreateClientScopeMappingsRealmRoles(ctx, token, realm, idOfClient, roles)
}

func (r *IdentityProvider) CreateClientScopeMappingsClientRoles(ctx context.Context, token, realm, idOfClient, idOfSelectedClient string, roles []gocloak.Role) error {
	return r.client.CreateClientScopeMappingsClientRoles(ctx, token, realm, idOfClient, idOfSelectedClient, roles)
}

func (r *IdentityProvider) CreateClientScopesScopeMappingsRealmRoles(ctx context.Context, token, realm, idOfClientScope string, roles []gocloak.Role) error {
	return r.client.CreateClientScopesScopeMappingsRealmRoles(ctx, token, realm, idOfClientScope, roles)
}

func (r *IdentityProvider) CreateClientScopesScopeMappingsClientRoles(ctx context.Context, token, realm, idOfClientScope, idOfClient string, roles []gocloak.Role) error {
	return r.client.CreateClientScopesScopeMappingsClientRoles(ctx, token, realm, idOfClientScope, idOfClient, roles)
}

func (r *IdentityProvider) CreateComponent(ctx context.Context, token, realm string, component gocloak.Component) (string, error) {
	return r.client.CreateComponent(ctx, token, realm, component)
}

func (r *IdentityProvider) GetComponent(ctx context.Context, token, realm, componentID string) (*gocloak.Component, error) {
	return r.client.GetComponent(ctx, token, realm, componentID)
}

func (r *IdentityProvider) GetComponentsWithParams(ctx context.Context, token, realm string, params gocloak.GetComponentsParams) ([]*gocloak.Component, error) {
	return r.client.GetComponentsWithParams(ctx, token, realm, params)
}

func (r *IdentityProvider) UpdateComponent(ctx context.Context, token, realm string, component gocloak.Component) error {
	return r.client.UpdateComponent(ctx, token, realm, component)
}

func (r *IdentityProvider) GetAuthenticationExecutions(ctx context.Context, token, realm, flow string) ([]*gocloak.ModifyAuthenticationExecutionRepresentation, error) {
	return r.client.GetAuthenticationExecutions(ctx, token, realm, flow)
}

func (r *IdentityProvider) CreateAuthenticationExecution(ctx context.Context, token, realm, flow string, execution gocloak.CreateAuthenticationExecutionRepresentation) error {
	return r.client.CreateAuthenticationExecution(ctx, token, realm, flow, execution)
}

func (r *IdentityProvider) UpdateAuthenticationExecution(ctx context.Context, token, realm, flow string, execution gocloak.ModifyAuthenticationExecutionRepresentation) error {
	return r.client.UpdateAuthenticationExecution(ctx, token, realm, flow, execution)
}

func (r *IdentityProvider) DeleteAuthenticationExecution(ctx context.Context, token, realm, executionID string) error {
	return r.client.DeleteAuthenticationExecution(ctx, token, realm, executionID)
}

func (r *IdentityProvider) CreateAuthenticationExecutionFlow(ctx context.Context, token, realm, flow string, executionFlow gocloak.CreateAuthenticationExecutionFlowRepresentation) error {
	return r.client.CreateAuthenticationExecutionFlow(ctx, token, realm, flow, executionFlow)
}

func (r *IdentityProvider) RegisterRequiredAction(ctx context.Context, token, realm string, requiredAction gocloak.RequiredActionProviderRepresentation) error {
	return r.client.RegisterRequiredAction(ctx, token, realm, requiredAction)
}

func (r *IdentityProvider) GetRequiredAction(ctx context.Context, token, realm, alias string) (*gocloak.RequiredActionProviderRepresentation, error) {
	return r.client.GetRequiredAction(ctx, token, realm, alias)
}

func (r *IdentityProvider) GetRequiredActions(ctx context.Context, token, realm string) ([]*gocloak.RequiredActionProviderRepresentation, error) {
	return r.client.GetRequiredActions(ctx, token, realm)
}

func (r *IdentityProvider) UpdateRequiredAction(ctx context.Context, token, realm string, requiredAction gocloak.RequiredActionProviderRepresentation) error {
	return r.client.UpdateRequiredAction(ctx, token, realm, requiredAction)
}

func (r *IdentityProvider) DeleteRequiredAction(ctx context.Context, token, realm, alias string) error {
	return r.client.DeleteRequiredAction(ctx, token, realm, alias)
}

func (r *IdentityProvider) GetIdentityProviderMappers(ctx context.Context, token, realm, alias string) ([]*gocloak.IdentityProviderMapper, error) {
	return r.client.GetIdentityProviderMappers(ctx, token, realm, alias)
}

func (r *IdentityProvider) GetIdentityProviderMapperByID(ctx context.Context, token, realm, alias, mapperID string) (*gocloak.IdentityProviderMapper, error) {
	return r.client.GetIdentityProviderMapperByID(ctx, token, realm, alias, mapperID)
}

func (r *IdentityProvider) UpdateIdentityProviderMapper(ctx context.Context, token, realm, alias string, mapper gocloak.IdentityProviderMapper) error {
	return r.client.UpdateIdentityProviderMapper(ctx, token, realm, alias, mapper)
}

func (r *IdentityProvider) DeleteIdentityProviderMapper(ctx context.Context, token, realm, alias, mapperID string) error {
	return r.client.DeleteIdentityProviderMapper(ctx, token, realm, alias, mapperID)
}

func (r *IdentityProvider) ExportIDPPublicBrokerConfig(ctx context.Context, token, realm, alias string) (*string, error) {
	return r.client.ExportIDPPublicBrokerConfig(ctx, token, realm, alias)
}

func (r *IdentityProvider) ImportIdentityProviderConfig(ctx context.Context, token, realm, fromURL, providerID string) (map[string]string, error) {
	return r.client.ImportIdentityProviderConfig(ctx, token, realm, fromURL, providerID)
}

func (r *IdentityProvider) ImportIdentityProviderConfigFromFile(ctx context.Context, token, realm, providerID, fileName string, fileBody io.Reader) (map[string]string, error) {
	return r.client.ImportIdentityProviderConfigFromFile(ctx, token, realm, providerID, fileName, fileBody)
}

func (r *IdentityProvider) GetAuthorizationPolicyAssociatedPolicies(ctx context.Context, token, realm, idOfClient, policyID string) ([]*gocloak.PolicyRepresentation, error) {
	return r.client.GetAuthorizationPolicyAssociatedPolicies(ctx, token, realm, idOfClient, policyID)
}

func (r *IdentityProvider) GetAuthorizationPolicyResources(ctx context.Context, token, realm, idOfClient, policyID string) ([]*gocloak.PolicyResourceRepresentation, error) {
	return r.client.GetAuthorizationPolicyResources(ctx, token, realm, idOfClient, policyID)
}

func (r *IdentityProvider) GetAuthorizationPolicyScopes(ctx context.Context, token, realm, idOfClient, policyID string) ([]*gocloak.PolicyScopeRepresentation, error) {
	return r.client.GetAuthorizationPolicyScopes(ctx, token, realm, idOfClient, policyID)
}

func (r *IdentityProvider) GetRequestingPartyToken(ctx context.Context, token, realm string, options gocloak.RequestingPartyTokenOptions) (*gocloak.JWT, error) {
	return r.client.GetRequestingPartyToken(ctx, token, realm, options)
}

func (r *IdentityProvider) GetRequestingPartyPermissions(ctx context.Context, token, realm string, options gocloak.RequestingPartyTokenOptions) (*[]gocloak.RequestingPartyPermission, error) {
	return r.client.GetRequestingPartyPermissions(ctx, token, realm, options)
}

func (r *IdentityProvider) GetRequestingPartyPermissionDecision(ctx context.Context, token, realm string, options gocloak.RequestingPartyTokenOptions) (*gocloak.RequestingPartyPermissionDecision, error) {
	return r.client.GetRequestingPartyPermissionDecision(ctx, token, realm, options)
}
