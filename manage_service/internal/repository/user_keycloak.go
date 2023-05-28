package repository

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/manage-service/internal/model"
	"github.com/JIeeiroSst/manage-service/pkg/log"

	keycloak "github.com/Nerzal/gocloak/v13"
)

type UserKeycloak interface {
	LoginAdmin(ctx context.Context, user model.Login) (*model.Token, error)
	GetTokenUser(ctx context.Context, realm string) (*model.TokenInfo, error)
	CreateUser(ctx context.Context, user model.CreateUser) error
	IntrospectToken(ctx context.Context, token model.IntrospectToken) (*[]keycloak.ResourcePermission, error)
	GetClients(ctx context.Context, user model.Client) ([]*keycloak.Client, error)
	Login(ctx context.Context, clientID, clientSecret, realm, username, password string) (*keycloak.JWT, error)
	LoginOtp(ctx context.Context, clientID, clientSecret, realm, username, password, totp string) (*keycloak.JWT, error)
	Logout(ctx context.Context, clientID, clientSecret, realm, refreshToken string) error
	LoginClient(ctx context.Context, clientID, clientSecret, realm string) (*keycloak.JWT, error)
	RefreshToken(ctx context.Context, refreshToken, clientID, clientSecret, realm string) (*keycloak.JWT, error)
	GetUserInfo(ctx context.Context, accessToken, realm string) (*keycloak.UserInfo, error)
	SetPassword(ctx context.Context, token, userID, realm, password string, temporary bool) error
	CreateGroup(ctx context.Context, accessToken, realm string, group keycloak.Group) (string, error)
	UpdateUser(ctx context.Context, accessToken, realm string, user keycloak.User) error
	UpdateGroup(ctx context.Context, accessToken, realm string, updatedGroup keycloak.Group) error
	UpdateRole(ctx context.Context, accessToken, realm, idOfClient string, role keycloak.Role) error
	UpdateClient(ctx context.Context, accessToken, realm string, updatedClient keycloak.Client) error
	UpdateClientScope(ctx context.Context, accessToken, realm string, scope keycloak.ClientScope) error
	DeleteUser(ctx context.Context, accessToken, realm, userID string) error
	DeleteComponent(ctx context.Context, accessToken, realm, componentID string) error
	DeleteGroup(ctx context.Context, accessToken, realm, groupID string) error
	DeleteClientRole(ctx context.Context, accessToken, realm, idOfClient, roleName string) error
	DeleteClientRoleFromUser(ctx context.Context, token, realm, idOfClient, userID string, roles []keycloak.Role) error
	DeleteClient(ctx context.Context, accessToken, realm, idOfClient string) error
	DeleteClientScope(ctx context.Context, accessToken, realm, scopeID string) error
	DeleteClientScopeMappingsRealmRoles(ctx context.Context, token, realm, idOfClient string, roles []keycloak.Role) error
	DeleteClientScopeMappingsClientRoles(ctx context.Context, token, realm, idOfClient, idOfSelectedClient string, roles []keycloak.Role) error
	DeleteClientScopesScopeMappingsRealmRoles(ctx context.Context, token, realm, idOfCLientScope string, roles []keycloak.Role) error
	DeleteClientScopesScopeMappingsClientRoles(ctx context.Context, token, realm, idOfClientScope, ifOfClient string, roles []keycloak.Role) error
	GetClient(ctx context.Context, accessToken, realm, idOfClient string) (*keycloak.Client, error)
	GetClientsDefaultScopes(ctx context.Context, token, realm, idOfClient string) ([]*keycloak.ClientScope, error)
	AddDefaultScopeToClient(ctx context.Context, token, realm, idOfClient, scopeID string) error
	RemoveDefaultScopeFromClient(ctx context.Context, token, realm, idOfClient, scopeID string) error
	GetClientsOptionalScopes(ctx context.Context, token, realm, idOfClient string) ([]*keycloak.ClientScope, error)
	AddOptionalScopeToClient(ctx context.Context, token, realm, idOfClient, scopeID string) error
	RemoveOptionalScopeFromClient(ctx context.Context, token, realm, idOfClient, scopeID string) error
	GetDefaultOptionalClientScopes(ctx context.Context, token, realm string) ([]*keycloak.ClientScope, error)
	GetDefaultDefaultClientScopes(ctx context.Context, token, realm string) ([]*keycloak.ClientScope, error)
	GetClientScope(ctx context.Context, token, realm, scopeID string) (*keycloak.ClientScope, error)
	GetClientScopes(ctx context.Context, token, realm string) ([]*keycloak.ClientScope, error)
	GetClientScopeMappings(ctx context.Context, token, realm, idOfClient string) (*keycloak.MappingsRepresentation, error)
	GetClientScopeMappingsRealmRoles(ctx context.Context, token, realm, idOfClient string) ([]*keycloak.Role, error)
	GetClientScopeMappingsRealmRolesAvailable(ctx context.Context, token, realm, idOfClient string) ([]*keycloak.Role, error)
	GetClientScopesScopeMappingsRealmRolesAvailable(ctx context.Context, token, realm, idOfClientScope string) ([]*keycloak.Role, error)
	GetClientScopesScopeMappingsClientRolesAvailable(ctx context.Context, token, realm, idOfClientScope, idOfClient string) ([]*keycloak.Role, error)
	GetClientScopeMappingsClientRoles(ctx context.Context, token, realm, idOfClient, idOfSelectedClient string) ([]*keycloak.Role, error)
	GetClientScopesScopeMappingsRealmRoles(ctx context.Context, token, realm, idOfClientScope string) ([]*keycloak.Role, error)
	GetClientScopesScopeMappingsClientRoles(ctx context.Context, token, realm, idOfClientScope, idOfClient string) ([]*keycloak.Role, error)
	GetClientScopeMappingsClientRolesAvailable(ctx context.Context, token, realm, idOfClient, idOfSelectedClient string) ([]*keycloak.Role, error)
	GetClientSecret(ctx context.Context, token, realm, idOfClient string) (*keycloak.CredentialRepresentation, error)
	GetClientServiceAccount(ctx context.Context, token, realm, idOfClient string) (*keycloak.User, error)
	RegenerateClientSecret(ctx context.Context, token, realm, idOfClient string) (*keycloak.CredentialRepresentation, error)
	GetKeyStoreConfig(ctx context.Context, accessToken, realm string) (*keycloak.KeyStoreConfig, error)
	GetUserByID(ctx context.Context, accessToken, realm, userID string) (*keycloak.User, error)
	GetUserCount(ctx context.Context, accessToken, realm string, params keycloak.GetUsersParams) (int, error)
	GetUsers(ctx context.Context, accessToken, realm string, params keycloak.GetUsersParams) ([]*keycloak.User, error)
	GetUserGroups(ctx context.Context, accessToken, realm, userID string, params keycloak.GetGroupsParams) ([]*keycloak.UserGroup, error)
	AddUserToGroup(ctx context.Context, token, realm, userID, groupID string) error
	DeleteUserFromGroup(ctx context.Context, token, realm, userID, groupID string) error
	GetComponents(ctx context.Context, accessToken, realm string) ([]*keycloak.Component, error)
	GetGroups(ctx context.Context, accessToken, realm string, params keycloak.GetGroupsParams) ([]*keycloak.Group, error)
	GetGroupsCount(ctx context.Context, token, realm string, params keycloak.GetGroupsParams) (int, error)
	GetGroup(ctx context.Context, accessToken, realm, groupID string) (*keycloak.Group, error)
	GetDefaultGroups(ctx context.Context, accessToken, realm string) ([]*keycloak.Group, error)
	AddDefaultGroup(ctx context.Context, accessToken, realm, groupID string) error
	RemoveDefaultGroup(ctx context.Context, accessToken, realm, groupID string) error
	GetGroupMembers(ctx context.Context, accessToken, realm, groupID string, params keycloak.GetGroupsParams) ([]*keycloak.User, error)
	GetRoleMappingByGroupID(ctx context.Context, accessToken, realm, groupID string) (*keycloak.MappingsRepresentation, error)
	GetRoleMappingByUserID(ctx context.Context, accessToken, realm, userID string) (*keycloak.MappingsRepresentation, error)
	GetClientRoles(ctx context.Context, accessToken, realm, idOfClient string, params keycloak.GetRoleParams) ([]*keycloak.Role, error)
	GetClientRole(ctx context.Context, token, realm, idOfClient, roleName string) (*keycloak.Role, error)
	GetClientRoleByID(ctx context.Context, accessToken, realm, roleID string) (*keycloak.Role, error)
	AddClientRoleComposite(ctx context.Context, token, realm, roleID string, roles []keycloak.Role) error
	DeleteClientRoleComposite(ctx context.Context, token, realm, roleID string, roles []keycloak.Role) error
	GetUsersByRoleName(ctx context.Context, token, realm, roleName string) ([]*keycloak.User, error)
	GetUsersByClientRoleName(ctx context.Context, token, realm, idOfClient, roleName string, params keycloak.GetUsersByRoleParams) ([]*keycloak.User, error)
	CreateClientProtocolMapper(ctx context.Context, token, realm, idOfClient string, mapper keycloak.ProtocolMapperRepresentation) (string, error)
	UpdateClientProtocolMapper(ctx context.Context, token, realm, idOfClient, mapperID string, mapper keycloak.ProtocolMapperRepresentation) error
	DeleteClientProtocolMapper(ctx context.Context, token, realm, idOfClient, mapperID string) error
	GetRealm(ctx context.Context, token, realm string) (*keycloak.RealmRepresentation, error)
	GetRealms(ctx context.Context, token string) ([]*keycloak.RealmRepresentation, error)
	CreateRealm(ctx context.Context, token string, realm keycloak.RealmRepresentation) (string, error)
	UpdateRealm(ctx context.Context, token string, realm keycloak.RealmRepresentation) error
	DeleteRealm(ctx context.Context, token, realm string) error
	ClearRealmCache(ctx context.Context, token, realm string) error
	ClearUserCache(ctx context.Context, token, realm string) error
	ClearKeysCache(ctx context.Context, token, realm string) error

	GetClientUserSessions(ctx context.Context, token, realm, idOfClient string) ([]*keycloak.UserSessionRepresentation, error)
	GetClientOfflineSessions(ctx context.Context, token, realm, idOfClient string) ([]*keycloak.UserSessionRepresentation, error)
	GetUserSessions(ctx context.Context, token, realm, userID string) ([]*keycloak.UserSessionRepresentation, error)
	GetUserOfflineSessionsForClient(ctx context.Context, token, realm, userID, idOfClient string) ([]*keycloak.UserSessionRepresentation, error)

	GetResource(ctx context.Context, token, realm, idOfClient, resourceID string) (*keycloak.ResourceRepresentation, error)
	GetResources(ctx context.Context, token, realm, idOfClient string, params keycloak.GetResourceParams) ([]*keycloak.ResourceRepresentation, error)
	CreateResource(ctx context.Context, token, realm, idOfClient string, resource keycloak.ResourceRepresentation) (*keycloak.ResourceRepresentation, error)
	UpdateResource(ctx context.Context, token, realm, idOfClient string, resource keycloak.ResourceRepresentation) error
	DeleteResource(ctx context.Context, token, realm, idOfClient, resourceID string) error

	GetResourceClient(ctx context.Context, token, realm, resourceID string) (*keycloak.ResourceRepresentation, error)
	GetResourcesClient(ctx context.Context, token, realm string, params keycloak.GetResourceParams) ([]*keycloak.ResourceRepresentation, error)
	CreateResourceClient(ctx context.Context, token, realm string, resource keycloak.ResourceRepresentation) (*keycloak.ResourceRepresentation, error)
	UpdateResourceClient(ctx context.Context, token, realm string, resource keycloak.ResourceRepresentation) error
	DeleteResourceClient(ctx context.Context, token, realm, resourceID string) error

	GetScope(ctx context.Context, token, realm, idOfClient, scopeID string) (*keycloak.ScopeRepresentation, error)
	GetScopes(ctx context.Context, token, realm, idOfClient string, params keycloak.GetScopeParams) ([]*keycloak.ScopeRepresentation, error)
	CreateScope(ctx context.Context, token, realm, idOfClient string, scope keycloak.ScopeRepresentation) (*keycloak.ScopeRepresentation, error)
	UpdateScope(ctx context.Context, token, realm, idOfClient string, resource keycloak.ScopeRepresentation) error
	DeleteScope(ctx context.Context, token, realm, idOfClient, scopeID string) error

	GetPolicy(ctx context.Context, token, realm, idOfClient, policyID string) (*keycloak.PolicyRepresentation, error)
	GetPolicies(ctx context.Context, token, realm, idOfClient string, params keycloak.GetPolicyParams) ([]*keycloak.PolicyRepresentation, error)
	CreatePolicy(ctx context.Context, token, realm, idOfClient string, policy keycloak.PolicyRepresentation) (*keycloak.PolicyRepresentation, error)
	UpdatePolicy(ctx context.Context, token, realm, idOfClient string, policy keycloak.PolicyRepresentation) error
	DeletePolicy(ctx context.Context, token, realm, idOfClient, policyID string) error

	GetResourcePolicy(ctx context.Context, token, realm, permissionID string) (*keycloak.ResourcePolicyRepresentation, error)
	GetResourcePolicies(ctx context.Context, token, realm string, params keycloak.GetResourcePoliciesParams) ([]*keycloak.ResourcePolicyRepresentation, error)
	CreateResourcePolicy(ctx context.Context, token, realm, resourceID string, policy keycloak.ResourcePolicyRepresentation) (*keycloak.ResourcePolicyRepresentation, error)
	UpdateResourcePolicy(ctx context.Context, token, realm, permissionID string, policy keycloak.ResourcePolicyRepresentation) error
	DeleteResourcePolicy(ctx context.Context, token, realm, permissionID string) error

	GetPermission(ctx context.Context, token, realm, idOfClient, permissionID string) (*keycloak.PermissionRepresentation, error)
	GetPermissions(ctx context.Context, token, realm, idOfClient string, params keycloak.GetPermissionParams) ([]*keycloak.PermissionRepresentation, error)
	GetPermissionResources(ctx context.Context, token, realm, idOfClient, permissionID string) ([]*keycloak.PermissionResource, error)
	GetPermissionScopes(ctx context.Context, token, realm, idOfClient, permissionID string) ([]*keycloak.PermissionScope, error)
	GetDependentPermissions(ctx context.Context, token, realm, idOfClient, policyID string) ([]*keycloak.PermissionRepresentation, error)
	CreatePermission(ctx context.Context, token, realm, idOfClient string, permission keycloak.PermissionRepresentation) (*keycloak.PermissionRepresentation, error)
	UpdatePermission(ctx context.Context, token, realm, idOfClient string, permission keycloak.PermissionRepresentation) error
	DeletePermission(ctx context.Context, token, realm, idOfClient, permissionID string) error

	CreatePermissionTicket(ctx context.Context, token, realm string, permissions []keycloak.CreatePermissionTicketParams) (*keycloak.PermissionTicketResponseRepresentation, error)
	GrantUserPermission(ctx context.Context, token, realm string, permission keycloak.PermissionGrantParams) (*keycloak.PermissionGrantResponseRepresentation, error)
	UpdateUserPermission(ctx context.Context, token, realm string, permission keycloak.PermissionGrantParams) (*keycloak.PermissionGrantResponseRepresentation, error)
	GetUserPermissions(ctx context.Context, token, realm string, params keycloak.GetUserPermissionParams) ([]*keycloak.PermissionGrantResponseRepresentation, error)
	DeleteUserPermission(ctx context.Context, token, realm, ticketID string) error

	GetCredentialRegistrators(ctx context.Context, token, realm string) ([]string, error)
	GetConfiguredUserStorageCredentialTypes(ctx context.Context, token, realm, userID string) ([]string, error)
	GetCredentials(ctx context.Context, token, realm, UserID string) ([]*keycloak.CredentialRepresentation, error)
	DeleteCredentials(ctx context.Context, token, realm, UserID, CredentialID string) error
	UpdateCredentialUserLabel(ctx context.Context, token, realm, userID, credentialID, userLabel string) error
	DisableAllCredentialsByType(ctx context.Context, token, realm, userID string, types []string) error
	MoveCredentialBehind(ctx context.Context, token, realm, userID, credentialID, newPreviousCredentialID string) error
	MoveCredentialToFirst(ctx context.Context, token, realm, userID, credentialID string) error

	GetAuthenticationFlows(ctx context.Context, token, realm string) ([]*keycloak.AuthenticationFlowRepresentation, error)
	GetAuthenticationFlow(ctx context.Context, token, realm string, authenticationFlowID string) (*keycloak.AuthenticationFlowRepresentation, error)
	CreateAuthenticationFlow(ctx context.Context, token, realm string, flow keycloak.AuthenticationFlowRepresentation) error
	UpdateAuthenticationFlow(ctx context.Context, token, realm string, flow keycloak.AuthenticationFlowRepresentation, authenticationFlowID string) (*keycloak.AuthenticationFlowRepresentation, error)
	DeleteAuthenticationFlow(ctx context.Context, token, realm, flowID string) error

	CreateIdentityProvider(ctx context.Context, token, realm string, providerRep keycloak.IdentityProviderRepresentation) (string, error)
	GetIdentityProvider(ctx context.Context, token, realm, alias string) (*keycloak.IdentityProviderRepresentation, error)
	GetIdentityProviders(ctx context.Context, token, realm string) ([]*keycloak.IdentityProviderRepresentation, error)
	UpdateIdentityProvider(ctx context.Context, token, realm, alias string, providerRep keycloak.IdentityProviderRepresentation) error
	DeleteIdentityProvider(ctx context.Context, token, realm, alias string) error

	CreateIdentityProviderMapper(ctx context.Context, token, realm, alias string, mapper keycloak.IdentityProviderMapper) (string, error)
	GetIdentityProviderMapper(ctx context.Context, token string, realm string, alias string, mapperID string) (*keycloak.IdentityProviderMapper, error)
	CreateUserFederatedIdentity(ctx context.Context, token, realm, userID, providerID string, federatedIdentityRep keycloak.FederatedIdentityRepresentation) error
	GetUserFederatedIdentities(ctx context.Context, token, realm, userID string) ([]*keycloak.FederatedIdentityRepresentation, error)
	DeleteUserFederatedIdentity(ctx context.Context, token, realm, userID, providerID string) error

	GetEvents(ctx context.Context, token string, realm string, params keycloak.GetEventsParams) ([]*keycloak.EventRepresentation, error)
}

type UserKeycloakRepo struct {
	client keycloak.GoCloak
}

func NewUserKeycloakRepo(client keycloak.GoCloak) UserKeycloak {
	return &UserKeycloakRepo{
		client: client,
	}
}

func (r *UserKeycloakRepo) LoginAdmin(ctx context.Context, user model.Login) (*model.Token, error) {
	token, err := r.client.LoginAdmin(ctx, user.User, user.Password, user.RealmName)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	log.Info(token)
	return &model.Token{
		Token: token.AccessToken,
	}, nil
}

func (r *UserKeycloakRepo) CreateUser(ctx context.Context, user model.CreateUser) error {
	userKeycloak := keycloak.User{
		FirstName: keycloak.StringP(user.FirstName),
		LastName:  keycloak.StringP(user.LastName),
		Email:     keycloak.StringP(user.Email),
		Enabled:   keycloak.BoolP(user.Enabled),
		Username:  keycloak.StringP(user.Username),
	}
	log.Info(userKeycloak)
	_, err := r.client.CreateUser(ctx, user.Token, user.Realm, userKeycloak)
	if err != nil {
		log.Error(err)
		return err

	}
	return nil
}

func (r *UserKeycloakRepo) IntrospectToken(ctx context.Context, token model.IntrospectToken) (*[]keycloak.ResourcePermission, error) {
	rptResult, err := r.client.RetrospectToken(ctx, token.Token, token.ClientID, token.ClientSecret, token.Realm)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	if !*rptResult.Active {
		log.Error("Token is not active")
		return nil, errors.New("Token is not active")
	}
	permissions := rptResult.Permissions
	return permissions, nil
}

func (r *UserKeycloakRepo) GetTokenUser(ctx context.Context, realm string) (*model.TokenInfo, error) {
	options := keycloak.TokenOptions{}
	client, err := r.client.GetToken(ctx, realm, options)
	if err != nil {
		return nil, err
	}
	return &model.TokenInfo{
		AccessToken:      client.AccessToken,
		RefreshToken:     client.RefreshToken,
		TokenType:        client.TokenType,
		ExpiresIn:        client.ExpiresIn,
		RefreshExpiresIn: client.RefreshExpiresIn,
		Scope:            client.Scope,
	}, nil
}

func (r *UserKeycloakRepo) GetClients(ctx context.Context, user model.Client) ([]*keycloak.Client, error) {
	clients, err := r.client.GetClients(
		ctx,
		user.Token,
		user.Realm,
		keycloak.GetClientsParams{
			ClientID: &user.ClientName,
		},
	)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	log.Info(clients)
	return clients, nil
}

func (r *UserKeycloakRepo) Login(ctx context.Context, clientID, clientSecret, realm, username, password string) (*keycloak.JWT, error) {
	token, err := r.client.Login(ctx, clientID, clientSecret, realm, username, password)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (r *UserKeycloakRepo) LoginOtp(ctx context.Context, clientID, clientSecret, realm, username, password, totp string) (*keycloak.JWT, error) {
	token, err := r.client.LoginOtp(ctx, clientID, clientSecret, realm, username, password, totp)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (r *UserKeycloakRepo) Logout(ctx context.Context, clientID, clientSecret, realm, refreshToken string) error {
	if err := r.client.Logout(ctx, clientID, clientSecret, realm, refreshToken); err != nil {
		return err
	}
	return nil
}

func (r *UserKeycloakRepo) LoginClient(ctx context.Context, clientID, clientSecret, realm string) (*keycloak.JWT, error) {
	token, err := r.client.LoginClient(ctx, clientID, clientSecret, realm)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (r *UserKeycloakRepo) RefreshToken(ctx context.Context, refreshToken, clientID, clientSecret, realm string) (*keycloak.JWT, error) {
	token, err := r.client.RefreshToken(ctx, refreshToken, clientID, clientSecret, realm)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (r *UserKeycloakRepo) GetUserInfo(ctx context.Context, accessToken, realm string) (*keycloak.UserInfo, error) {
	userInfo, err := r.client.GetUserInfo(ctx, accessToken, realm)
	if err != nil {
		return nil, err
	}
	return userInfo, nil
}

func (r *UserKeycloakRepo) SetPassword(ctx context.Context, token, userID, realm, password string, temporary bool) error {
	if err := r.client.SetPassword(ctx, token, userID, realm, password, temporary); err != nil {
		return err
	}
	return nil
}

func (r *UserKeycloakRepo) CreateGroup(ctx context.Context, accessToken, realm string, group keycloak.Group) (string, error) {
	msg, err := r.client.CreateGroup(ctx, accessToken, realm, group)
	if err != nil {
		return "", err
	}
	return msg, nil
}

func (r *UserKeycloakRepo) UpdateUser(ctx context.Context, accessToken, realm string, user keycloak.User) error {
	if err := r.client.UpdateUser(ctx, accessToken, realm, user); err != nil {
		return err
	}
	return nil
}

func (r *UserKeycloakRepo) UpdateGroup(ctx context.Context, accessToken, realm string, updatedGroup keycloak.Group) error {
	if err := r.client.UpdateGroup(ctx, accessToken, realm, updatedGroup); err != nil {
		return err
	}
	return nil
}

func (r *UserKeycloakRepo) UpdateRole(ctx context.Context, accessToken, realm, idOfClient string, role keycloak.Role) error {
	if err := r.client.UpdateRole(ctx, accessToken, realm, idOfClient, role); err != nil {
		return err
	}
	return nil
}

func (r *UserKeycloakRepo) UpdateClient(ctx context.Context, accessToken, realm string, updatedClient keycloak.Client) error {
	if err := r.client.UpdateClient(ctx, accessToken, realm, updatedClient); err != nil {
		return err
	}
	return nil
}

func (r *UserKeycloakRepo) UpdateClientScope(ctx context.Context, accessToken, realm string, scope keycloak.ClientScope) error {
	if err := r.client.UpdateClientScope(ctx, accessToken, realm, scope); err != nil {
		return err
	}
	return nil
}

func (r *UserKeycloakRepo) DeleteUser(ctx context.Context, accessToken, realm, userID string) error {
	if err := r.client.DeleteClient(ctx, accessToken, realm, userID); err != nil {
		return err
	}
	return nil
}

func (r *UserKeycloakRepo) DeleteComponent(ctx context.Context, accessToken, realm, componentID string) error {
	if err := r.client.DeleteComponent(ctx, accessToken, realm, componentID); err != nil {
		return err
	}
	return nil
}

func (r *UserKeycloakRepo) DeleteGroup(ctx context.Context, accessToken, realm, groupID string) error {
	if err := r.client.DeleteGroup(ctx, accessToken, realm, groupID); err != nil {
		return err
	}
	return nil
}

func (r *UserKeycloakRepo) DeleteClientRole(ctx context.Context, accessToken, realm, idOfClient, roleName string) error {
	if err := r.client.DeleteClientRole(ctx, accessToken, realm, idOfClient, roleName); err != nil {
		return err
	}
	return nil
}

func (r *UserKeycloakRepo) DeleteClientRoleFromUser(ctx context.Context, token, realm, idOfClient, userID string, roles []keycloak.Role) error {
	if err := r.client.DeleteClientRoleFromUser(ctx, token, realm, idOfClient, userID, roles); err != nil {
		return err
	}
	return nil
}

func (r *UserKeycloakRepo) DeleteClient(ctx context.Context, accessToken, realm, idOfClient string) error {
	if err := r.client.DeleteClient(ctx, accessToken, realm, idOfClient); err != nil {
		return err
	}
	return nil
}

func (r *UserKeycloakRepo) DeleteClientScope(ctx context.Context, accessToken, realm, scopeID string) error {
	if err := r.client.DeleteClientScope(ctx, accessToken, realm, scopeID); err != nil {
		return err
	}
	return nil
}

func (r *UserKeycloakRepo) DeleteClientScopeMappingsRealmRoles(ctx context.Context, token, realm, idOfClient string, roles []keycloak.Role) error {
	if err := r.client.DeleteClientScopeMappingsRealmRoles(ctx, token, realm, idOfClient, roles); err != nil {
		return err
	}
	return nil
}

func (r *UserKeycloakRepo) DeleteClientScopeMappingsClientRoles(ctx context.Context, token, realm, idOfClient, idOfSelectedClient string, roles []keycloak.Role) error {
	if err := r.client.DeleteClientScopeMappingsClientRoles(ctx, token, realm, idOfClient, idOfSelectedClient, roles); err != nil {
		return err
	}
	return nil
}

func (r *UserKeycloakRepo) DeleteClientScopesScopeMappingsRealmRoles(ctx context.Context, token, realm, idOfCLientScope string, roles []keycloak.Role) error {
	if err := r.client.DeleteClientScopesScopeMappingsRealmRoles(ctx, token, realm, idOfCLientScope, roles); err != nil {
		return err
	}
	return nil
}

func (r *UserKeycloakRepo) DeleteClientScopesScopeMappingsClientRoles(ctx context.Context, token, realm, idOfClientScope, ifOfClient string, roles []keycloak.Role) error {
	if err := r.client.DeleteClientScopesScopeMappingsClientRoles(ctx, token, realm, idOfClientScope, ifOfClient, roles); err != nil {
		return err
	}
	return nil
}

func (r *UserKeycloakRepo) GetClient(ctx context.Context, accessToken, realm, idOfClient string) (*keycloak.Client, error) {
	client, err := r.client.GetClient(ctx, accessToken, realm, idOfClient)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (r *UserKeycloakRepo) GetClientsDefaultScopes(ctx context.Context, token, realm, idOfClient string) ([]*keycloak.ClientScope, error) {
	clientScope, err := r.client.GetClientsDefaultScopes(ctx, token, realm, idOfClient)
	if err != nil {
		return nil, err
	}
	return clientScope, nil
}

func (r *UserKeycloakRepo) AddDefaultScopeToClient(ctx context.Context, token, realm, idOfClient, scopeID string) error {
	if err := r.client.AddDefaultScopeToClient(ctx, token, realm, idOfClient, scopeID); err != nil {
		return err
	}
	return nil
}

func (r *UserKeycloakRepo) RemoveDefaultScopeFromClient(ctx context.Context, token, realm, idOfClient, scopeID string) error {
	if err := r.client.RemoveDefaultScopeFromClient(ctx, token, realm, idOfClient, scopeID); err != nil {
		return err
	}
	return nil
}

func (r *UserKeycloakRepo) GetClientsOptionalScopes(ctx context.Context, token, realm, idOfClient string) ([]*keycloak.ClientScope, error) {
	clientScope, err := r.client.GetClientsOptionalScopes(ctx, token, realm, idOfClient)
	if err != nil {
		return nil, err
	}
	return clientScope, nil
}

func (r *UserKeycloakRepo) AddOptionalScopeToClient(ctx context.Context, token, realm, idOfClient, scopeID string) error {
	if err := r.client.AddOptionalScopeToClient(ctx, token, realm, idOfClient, scopeID); err != nil {
		return err
	}
	return nil
}

func (r *UserKeycloakRepo) RemoveOptionalScopeFromClient(ctx context.Context, token, realm, idOfClient, scopeID string) error {
	if err := r.client.RemoveOptionalScopeFromClient(ctx, token, realm, idOfClient, scopeID); err != nil {
		return err
	}
	return nil
}

func (r *UserKeycloakRepo) GetDefaultOptionalClientScopes(ctx context.Context, token, realm string) ([]*keycloak.ClientScope, error) {
	clientScope, err := r.client.GetDefaultOptionalClientScopes(ctx, token, realm)
	if err != nil {
		return nil, err
	}
	return clientScope, nil
}

func (r *UserKeycloakRepo) GetDefaultDefaultClientScopes(ctx context.Context, token, realm string) ([]*keycloak.ClientScope, error) {
	clientScope, err := r.client.GetDefaultDefaultClientScopes(ctx, token, realm)
	if err != nil {
		return nil, err
	}
	return clientScope, nil
}

func (r *UserKeycloakRepo) GetClientScope(ctx context.Context, token, realm, scopeID string) (*keycloak.ClientScope, error) {
	clientScope, err := r.client.GetClientScope(ctx, token, realm, scopeID)
	if err != nil {
		return nil, err
	}
	return clientScope, nil
}

func (r *UserKeycloakRepo) GetClientScopes(ctx context.Context, token, realm string) ([]*keycloak.ClientScope, error) {
	clientScope, err := r.client.GetClientScopes(ctx, token, realm)
	if err != nil {
		return nil, err
	}
	return clientScope, nil
}

func (r *UserKeycloakRepo) GetClientScopeMappings(ctx context.Context, token, realm, idOfClient string) (*keycloak.MappingsRepresentation, error) {
	mappingsRepresentation, err := r.client.GetClientScopeMappings(ctx, token, realm, idOfClient)
	if err != nil {
		return nil, err
	}
	return mappingsRepresentation, nil
}

func (r *UserKeycloakRepo) GetClientScopeMappingsRealmRoles(ctx context.Context, token, realm, idOfClient string) ([]*keycloak.Role, error) {
	role, err := r.client.GetClientScopeMappingsRealmRoles(ctx, token, realm, idOfClient)
	if err != nil {
		return nil, err
	}
	return role, nil
}

func (r *UserKeycloakRepo) GetClientScopeMappingsRealmRolesAvailable(ctx context.Context, token, realm, idOfClient string) ([]*keycloak.Role, error) {
	role, err := r.client.GetClientScopeMappingsRealmRolesAvailable(ctx, token, realm, idOfClient)
	if err != nil {
		return nil, err
	}
	return role, nil
}

func (r *UserKeycloakRepo) GetClientScopesScopeMappingsRealmRolesAvailable(ctx context.Context, token, realm, idOfClientScope string) ([]*keycloak.Role, error) {
	role, err := r.client.GetClientScopesScopeMappingsRealmRolesAvailable(ctx, token, realm, idOfClientScope)
	if err != nil {
		return nil, err
	}
	return role, nil
}

func (r *UserKeycloakRepo) GetClientScopesScopeMappingsClientRolesAvailable(ctx context.Context, token, realm, idOfClientScope, idOfClient string) ([]*keycloak.Role, error) {
	role, err := r.client.GetClientScopesScopeMappingsClientRolesAvailable(ctx, token, realm, idOfClientScope, idOfClient)
	if err != nil {
		return nil, err
	}
	return role, nil
}

func (r *UserKeycloakRepo) GetClientScopeMappingsClientRoles(ctx context.Context, token, realm, idOfClient, idOfSelectedClient string) ([]*keycloak.Role, error) {
	role, err := r.client.GetClientScopeMappingsClientRoles(ctx, token, realm, idOfClient, idOfSelectedClient)
	if err != nil {
		return nil, err
	}
	return role, nil
}

func (r *UserKeycloakRepo) GetClientScopesScopeMappingsRealmRoles(ctx context.Context, token, realm, idOfClientScope string) ([]*keycloak.Role, error) {
	role, err := r.client.GetClientScopesScopeMappingsRealmRoles(ctx, token, realm, idOfClientScope)
	if err != nil {
		return nil, err
	}
	return role, nil
}

func (r *UserKeycloakRepo) GetClientScopesScopeMappingsClientRoles(ctx context.Context, token, realm, idOfClientScope, idOfClient string) ([]*keycloak.Role, error) {
	role, err := r.client.GetClientScopesScopeMappingsClientRoles(ctx, token, realm, idOfClientScope, idOfClient)
	if err != nil {
		return nil, err
	}
	return role, nil
}

func (r *UserKeycloakRepo) GetClientScopeMappingsClientRolesAvailable(ctx context.Context, token, realm, idOfClient, idOfSelectedClient string) ([]*keycloak.Role, error) {
	role, err := r.client.GetClientScopeMappingsClientRolesAvailable(ctx, token, realm, idOfClient, idOfSelectedClient)
	if err != nil {
		return nil, err
	}
	return role, nil
}

func (r *UserKeycloakRepo) GetClientSecret(ctx context.Context, token, realm, idOfClient string) (*keycloak.CredentialRepresentation, error) {
	clientSecret, err := r.client.GetClientSecret(ctx, token, realm, idOfClient)
	if err != nil {
		return nil, err
	}
	return clientSecret, nil
}

func (r *UserKeycloakRepo) GetClientServiceAccount(ctx context.Context, token, realm, idOfClient string) (*keycloak.User, error) {
	user, err := r.client.GetClientServiceAccount(ctx, token, realm, idOfClient)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserKeycloakRepo) RegenerateClientSecret(ctx context.Context, token, realm, idOfClient string) (*keycloak.CredentialRepresentation, error) {
	credentialRepresentation, err := r.client.RegenerateClientSecret(ctx, token, realm, idOfClient)
	if err != nil {
		return nil, err
	}
	return credentialRepresentation, nil
}

func (r *UserKeycloakRepo) GetKeyStoreConfig(ctx context.Context, accessToken, realm string) (*keycloak.KeyStoreConfig, error) {
	keyconfig, err := r.client.GetKeyStoreConfig(ctx, accessToken, realm)
	if err != nil {
		return nil, err
	}
	return keyconfig, nil
}

func (r *UserKeycloakRepo) GetUserByID(ctx context.Context, accessToken, realm, userID string) (*keycloak.User, error) {
	user, err := r.client.GetUserByID(ctx, accessToken, realm, userID)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserKeycloakRepo) GetUserCount(ctx context.Context, accessToken, realm string, params keycloak.GetUsersParams) (int, error) {
	count, err := r.client.GetUserCount(ctx, accessToken, realm, params)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *UserKeycloakRepo) GetUsers(ctx context.Context, accessToken, realm string, params keycloak.GetUsersParams) ([]*keycloak.User, error)

func (r *UserKeycloakRepo) GetUserGroups(ctx context.Context, accessToken, realm, userID string, params keycloak.GetGroupsParams) ([]*keycloak.UserGroup, error)

func (r *UserKeycloakRepo) AddUserToGroup(ctx context.Context, token, realm, userID, groupID string) error

func (r *UserKeycloakRepo) DeleteUserFromGroup(ctx context.Context, token, realm, userID, groupID string) error

func (r *UserKeycloakRepo) GetComponents(ctx context.Context, accessToken, realm string) ([]*keycloak.Component, error)

func (r *UserKeycloakRepo) GetGroups(ctx context.Context, accessToken, realm string, params keycloak.GetGroupsParams) ([]*keycloak.Group, error)

func (r *UserKeycloakRepo) GetGroupsCount(ctx context.Context, token, realm string, params keycloak.GetGroupsParams) (int, error)

func (r *UserKeycloakRepo) GetGroup(ctx context.Context, accessToken, realm, groupID string) (*keycloak.Group, error)

func (r *UserKeycloakRepo) GetDefaultGroups(ctx context.Context, accessToken, realm string) ([]*keycloak.Group, error)

func (r *UserKeycloakRepo) AddDefaultGroup(ctx context.Context, accessToken, realm, groupID string) error

func (r *UserKeycloakRepo) RemoveDefaultGroup(ctx context.Context, accessToken, realm, groupID string) error

func (r *UserKeycloakRepo) GetGroupMembers(ctx context.Context, accessToken, realm, groupID string, params keycloak.GetGroupsParams) ([]*keycloak.User, error)

func (r *UserKeycloakRepo) GetRoleMappingByGroupID(ctx context.Context, accessToken, realm, groupID string) (*keycloak.MappingsRepresentation, error)

func (r *UserKeycloakRepo) GetRoleMappingByUserID(ctx context.Context, accessToken, realm, userID string) (*keycloak.MappingsRepresentation, error)

func (r *UserKeycloakRepo) GetClientRoles(ctx context.Context, accessToken, realm, idOfClient string, params keycloak.GetRoleParams) ([]*keycloak.Role, error)

func (r *UserKeycloakRepo) GetClientRole(ctx context.Context, token, realm, idOfClient, roleName string) (*keycloak.Role, error)

func (r *UserKeycloakRepo) GetClientRoleByID(ctx context.Context, accessToken, realm, roleID string) (*keycloak.Role, error)

func (r *UserKeycloakRepo) AddClientRoleComposite(ctx context.Context, token, realm, roleID string, roles []keycloak.Role) error

func (r *UserKeycloakRepo) DeleteClientRoleComposite(ctx context.Context, token, realm, roleID string, roles []keycloak.Role) error

func (r *UserKeycloakRepo) GetUsersByRoleName(ctx context.Context, token, realm, roleName string) ([]*keycloak.User, error)

func (r *UserKeycloakRepo) GetUsersByClientRoleName(ctx context.Context, token, realm, idOfClient, roleName string, params keycloak.GetUsersByRoleParams) ([]*keycloak.User, error)

func (r *UserKeycloakRepo) CreateClientProtocolMapper(ctx context.Context, token, realm, idOfClient string, mapper keycloak.ProtocolMapperRepresentation) (string, error)

func (r *UserKeycloakRepo) UpdateClientProtocolMapper(ctx context.Context, token, realm, idOfClient, mapperID string, mapper keycloak.ProtocolMapperRepresentation) error

func (r *UserKeycloakRepo) DeleteClientProtocolMapper(ctx context.Context, token, realm, idOfClient, mapperID string) error

func (r *UserKeycloakRepo) GetRealm(ctx context.Context, token, realm string) (*keycloak.RealmRepresentation, error)

func (r *UserKeycloakRepo) GetRealms(ctx context.Context, token string) ([]*keycloak.RealmRepresentation, error)

func (r *UserKeycloakRepo) CreateRealm(ctx context.Context, token string, realm keycloak.RealmRepresentation) (string, error)

func (r *UserKeycloakRepo) UpdateRealm(ctx context.Context, token string, realm keycloak.RealmRepresentation) error

func (r *UserKeycloakRepo) DeleteRealm(ctx context.Context, token, realm string) error

func (r *UserKeycloakRepo) ClearRealmCache(ctx context.Context, token, realm string) error

func (r *UserKeycloakRepo) ClearUserCache(ctx context.Context, token, realm string) error

func (r *UserKeycloakRepo) ClearKeysCache(ctx context.Context, token, realm string) error

func (r *UserKeycloakRepo) GetClientUserSessions(ctx context.Context, token, realm, idOfClient string) ([]*keycloak.UserSessionRepresentation, error)
func (r *UserKeycloakRepo) GetClientOfflineSessions(ctx context.Context, token, realm, idOfClient string) ([]*keycloak.UserSessionRepresentation, error)
func (r *UserKeycloakRepo) GetUserSessions(ctx context.Context, token, realm, userID string) ([]*keycloak.UserSessionRepresentation, error)
func (r *UserKeycloakRepo) GetUserOfflineSessionsForClient(ctx context.Context, token, realm, userID, idOfClient string) ([]*keycloak.UserSessionRepresentation, error)

func (r *UserKeycloakRepo) GetResource(ctx context.Context, token, realm, idOfClient, resourceID string) (*keycloak.ResourceRepresentation, error)
func (r *UserKeycloakRepo) GetResources(ctx context.Context, token, realm, idOfClient string, params keycloak.GetResourceParams) ([]*keycloak.ResourceRepresentation, error)
func (r *UserKeycloakRepo) CreateResource(ctx context.Context, token, realm, idOfClient string, resource keycloak.ResourceRepresentation) (*keycloak.ResourceRepresentation, error)
func (r *UserKeycloakRepo) UpdateResource(ctx context.Context, token, realm, idOfClient string, resource keycloak.ResourceRepresentation) error
func (r *UserKeycloakRepo) DeleteResource(ctx context.Context, token, realm, idOfClient, resourceID string) error

func (r *UserKeycloakRepo) GetResourceClient(ctx context.Context, token, realm, resourceID string) (*keycloak.ResourceRepresentation, error)
func (r *UserKeycloakRepo) GetResourcesClient(ctx context.Context, token, realm string, params keycloak.GetResourceParams) ([]*keycloak.ResourceRepresentation, error)
func (r *UserKeycloakRepo) CreateResourceClient(ctx context.Context, token, realm string, resource keycloak.ResourceRepresentation) (*keycloak.ResourceRepresentation, error)
func (r *UserKeycloakRepo) UpdateResourceClient(ctx context.Context, token, realm string, resource keycloak.ResourceRepresentation) error
func (r *UserKeycloakRepo) DeleteResourceClient(ctx context.Context, token, realm, resourceID string) error

func (r *UserKeycloakRepo) GetScope(ctx context.Context, token, realm, idOfClient, scopeID string) (*keycloak.ScopeRepresentation, error)
func (r *UserKeycloakRepo) GetScopes(ctx context.Context, token, realm, idOfClient string, params keycloak.GetScopeParams) ([]*keycloak.ScopeRepresentation, error)
func (r *UserKeycloakRepo) CreateScope(ctx context.Context, token, realm, idOfClient string, scope keycloak.ScopeRepresentation) (*keycloak.ScopeRepresentation, error)
func (r *UserKeycloakRepo) UpdateScope(ctx context.Context, token, realm, idOfClient string, resource keycloak.ScopeRepresentation) error
func (r *UserKeycloakRepo) DeleteScope(ctx context.Context, token, realm, idOfClient, scopeID string) error

func (r *UserKeycloakRepo) GetPolicy(ctx context.Context, token, realm, idOfClient, policyID string) (*keycloak.PolicyRepresentation, error)
func (r *UserKeycloakRepo) GetPolicies(ctx context.Context, token, realm, idOfClient string, params keycloak.GetPolicyParams) ([]*keycloak.PolicyRepresentation, error)
func (r *UserKeycloakRepo) CreatePolicy(ctx context.Context, token, realm, idOfClient string, policy keycloak.PolicyRepresentation) (*keycloak.PolicyRepresentation, error)
func (r *UserKeycloakRepo) UpdatePolicy(ctx context.Context, token, realm, idOfClient string, policy keycloak.PolicyRepresentation) error
func (r *UserKeycloakRepo) DeletePolicy(ctx context.Context, token, realm, idOfClient, policyID string) error

func (r *UserKeycloakRepo) GetResourcePolicy(ctx context.Context, token, realm, permissionID string) (*keycloak.ResourcePolicyRepresentation, error)
func (r *UserKeycloakRepo) GetResourcePolicies(ctx context.Context, token, realm string, params keycloak.GetResourcePoliciesParams) ([]*keycloak.ResourcePolicyRepresentation, error)
func (r *UserKeycloakRepo) CreateResourcePolicy(ctx context.Context, token, realm, resourceID string, policy keycloak.ResourcePolicyRepresentation) (*keycloak.ResourcePolicyRepresentation, error)
func (r *UserKeycloakRepo) UpdateResourcePolicy(ctx context.Context, token, realm, permissionID string, policy keycloak.ResourcePolicyRepresentation) error
func (r *UserKeycloakRepo) DeleteResourcePolicy(ctx context.Context, token, realm, permissionID string) error

func (r *UserKeycloakRepo) GetPermission(ctx context.Context, token, realm, idOfClient, permissionID string) (*keycloak.PermissionRepresentation, error)
func (r *UserKeycloakRepo) GetPermissions(ctx context.Context, token, realm, idOfClient string, params keycloak.GetPermissionParams) ([]*keycloak.PermissionRepresentation, error)
func (r *UserKeycloakRepo) GetPermissionResources(ctx context.Context, token, realm, idOfClient, permissionID string) ([]*keycloak.PermissionResource, error)
func (r *UserKeycloakRepo) GetPermissionScopes(ctx context.Context, token, realm, idOfClient, permissionID string) ([]*keycloak.PermissionScope, error)
func (r *UserKeycloakRepo) GetDependentPermissions(ctx context.Context, token, realm, idOfClient, policyID string) ([]*keycloak.PermissionRepresentation, error)
func (r *UserKeycloakRepo) CreatePermission(ctx context.Context, token, realm, idOfClient string, permission keycloak.PermissionRepresentation) (*keycloak.PermissionRepresentation, error)
func (r *UserKeycloakRepo) UpdatePermission(ctx context.Context, token, realm, idOfClient string, permission keycloak.PermissionRepresentation) error
func (r *UserKeycloakRepo) DeletePermission(ctx context.Context, token, realm, idOfClient, permissionID string) error

func (r *UserKeycloakRepo) CreatePermissionTicket(ctx context.Context, token, realm string, permissions []keycloak.CreatePermissionTicketParams) (*keycloak.PermissionTicketResponseRepresentation, error)
func (r *UserKeycloakRepo) GrantUserPermission(ctx context.Context, token, realm string, permission keycloak.PermissionGrantParams) (*keycloak.PermissionGrantResponseRepresentation, error)
func (r *UserKeycloakRepo) UpdateUserPermission(ctx context.Context, token, realm string, permission keycloak.PermissionGrantParams) (*keycloak.PermissionGrantResponseRepresentation, error)
func (r *UserKeycloakRepo) GetUserPermissions(ctx context.Context, token, realm string, params keycloak.GetUserPermissionParams) ([]*keycloak.PermissionGrantResponseRepresentation, error)
func (r *UserKeycloakRepo) DeleteUserPermission(ctx context.Context, token, realm, ticketID string) error

func (r *UserKeycloakRepo) GetCredentialRegistrators(ctx context.Context, token, realm string) ([]string, error)
func (r *UserKeycloakRepo) GetConfiguredUserStorageCredentialTypes(ctx context.Context, token, realm, userID string) ([]string, error)
func (r *UserKeycloakRepo) GetCredentials(ctx context.Context, token, realm, UserID string) ([]*keycloak.CredentialRepresentation, error)
func (r *UserKeycloakRepo) DeleteCredentials(ctx context.Context, token, realm, UserID, CredentialID string) error
func (r *UserKeycloakRepo) UpdateCredentialUserLabel(ctx context.Context, token, realm, userID, credentialID, userLabel string) error
func (r *UserKeycloakRepo) DisableAllCredentialsByType(ctx context.Context, token, realm, userID string, types []string) error
func (r *UserKeycloakRepo) MoveCredentialBehind(ctx context.Context, token, realm, userID, credentialID, newPreviousCredentialID string) error
func (r *UserKeycloakRepo) MoveCredentialToFirst(ctx context.Context, token, realm, userID, credentialID string) error

func (r *UserKeycloakRepo) GetAuthenticationFlows(ctx context.Context, token, realm string) ([]*keycloak.AuthenticationFlowRepresentation, error)
func (r *UserKeycloakRepo) GetAuthenticationFlow(ctx context.Context, token, realm string, authenticationFlowID string) (*keycloak.AuthenticationFlowRepresentation, error)
func (r *UserKeycloakRepo) CreateAuthenticationFlow(ctx context.Context, token, realm string, flow keycloak.AuthenticationFlowRepresentation) error
func (r *UserKeycloakRepo) UpdateAuthenticationFlow(ctx context.Context, token, realm string, flow keycloak.AuthenticationFlowRepresentation, authenticationFlowID string) (*keycloak.AuthenticationFlowRepresentation, error)
func (r *UserKeycloakRepo) DeleteAuthenticationFlow(ctx context.Context, token, realm, flowID string) error

func (r *UserKeycloakRepo) CreateIdentityProvider(ctx context.Context, token, realm string, providerRep keycloak.IdentityProviderRepresentation) (string, error)
func (r *UserKeycloakRepo) GetIdentityProvider(ctx context.Context, token, realm, alias string) (*keycloak.IdentityProviderRepresentation, error)
func (r *UserKeycloakRepo) GetIdentityProviders(ctx context.Context, token, realm string) ([]*keycloak.IdentityProviderRepresentation, error)
func (r *UserKeycloakRepo) UpdateIdentityProvider(ctx context.Context, token, realm, alias string, providerRep keycloak.IdentityProviderRepresentation) error
func (r *UserKeycloakRepo) DeleteIdentityProvider(ctx context.Context, token, realm, alias string) error

func (r *UserKeycloakRepo) CreateIdentityProviderMapper(ctx context.Context, token, realm, alias string, mapper keycloak.IdentityProviderMapper) (string, error)
func (r *UserKeycloakRepo) GetIdentityProviderMapper(ctx context.Context, token string, realm string, alias string, mapperID string) (*keycloak.IdentityProviderMapper, error)
func (r *UserKeycloakRepo) CreateUserFederatedIdentity(ctx context.Context, token, realm, userID, providerID string, federatedIdentityRep keycloak.FederatedIdentityRepresentation) error
func (r *UserKeycloakRepo) GetUserFederatedIdentities(ctx context.Context, token, realm, userID string) ([]*keycloak.FederatedIdentityRepresentation, error)
func (r *UserKeycloakRepo) DeleteUserFederatedIdentity(ctx context.Context, token, realm, userID, providerID string) error

func (r *UserKeycloakRepo) GetEvents(ctx context.Context, token string, realm string, params keycloak.GetEventsParams) ([]*keycloak.EventRepresentation, error)
