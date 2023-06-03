package usecase

import (
	"context"

	"github.com/JIeeiroSst/manage-service/internal/dto"
	"github.com/JIeeiroSst/manage-service/internal/model"
	"github.com/JIeeiroSst/manage-service/internal/repository"
	keycloak "github.com/Nerzal/gocloak/v13"
)

type UserKeyclock interface {
	LoginAdmin(ctx context.Context, user dto.Login) (*dto.Token, error)
	GetTokenUser(ctx context.Context, realm string) (*dto.TokenInfo, error)
	CreateUser(ctx context.Context, user dto.CreateUser) error

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
	GetUserGroups(ctx context.Context, accessToken, realm, userID string, params keycloak.GetGroupsParams) ([]*keycloak.Group, error)
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
	GetUsersByRoleName(ctx context.Context, token, realm, roleName string, roles keycloak.GetUsersByRoleParams) ([]*keycloak.User, error)
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

type UserKeycloakUsecase struct {
	UserKeycloakRepo repository.UserKeycloak
}

func NewUserKeycloakUsecase(UserKeycloakRepo repository.UserKeycloak) *UserKeycloakUsecase {
	return &UserKeycloakUsecase{
		UserKeycloakRepo: UserKeycloakRepo,
	}
}

func (u *UserKeycloakUsecase) LoginAdmin(ctx context.Context, user dto.Login) (*dto.Token, error) {
	userModel := model.Login{}
	token, err := u.UserKeycloakRepo.LoginAdmin(ctx, userModel)
	if err != nil {
		return nil, err
	}
	return &dto.Token{
		Token: token.Token,
	}, nil
}

func (u *UserKeycloakUsecase) GetTokenUser(ctx context.Context, realm string) (*dto.TokenInfo, error) {
	token, err := u.UserKeycloakRepo.GetTokenUser(ctx, realm)
	if err != nil {
		return nil, err
	}
	return &dto.TokenInfo{
		AccessToken:      token.AccessToken,
		RefreshToken:     token.RefreshToken,
		TokenType:        token.TokenType,
		ExpiresIn:        token.ExpiresIn,
		RefreshExpiresIn: token.RefreshExpiresIn,
		Scope:            token.Scope,
	}, nil
}

func (u *UserKeycloakUsecase) CreateUser(ctx context.Context, user dto.CreateUser) error {
	userModel := model.CreateUser{
		Token:     user.Token,
		Realm:     user.Realm,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.LastName,
		Enabled:   user.Enabled,
		Username:  user.Username,
	}
	if err := u.UserKeycloakRepo.CreateUser(ctx, userModel); err != nil {
		return err
	}
	return nil
}

func (u *UserKeycloakUsecase) IntrospectToken(ctx context.Context, token model.IntrospectToken) (*[]keycloak.ResourcePermission, error) {
	resourcePermission, err := u.UserKeycloakRepo.IntrospectToken(ctx, token)
	if err != nil {
		return nil, err
	}
	return resourcePermission, nil
}

func (u *UserKeycloakUsecase) GetClients(ctx context.Context, user model.Client) ([]*keycloak.Client, error) {
	client, err := u.UserKeycloakRepo.GetClients(ctx, user)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (u *UserKeycloakUsecase) Login(ctx context.Context, clientID, clientSecret, realm, username, password string) (*keycloak.JWT, error) {
	token, err := u.UserKeycloakRepo.Login(ctx, clientID, clientSecret, realm, username, password)
	if err != nil {
		return nil, err
	}
	return token, nil
}
func (u *UserKeycloakUsecase) LoginOtp(ctx context.Context, clientID, clientSecret, realm, username, password, totp string) (*keycloak.JWT, error) {
	token, err := u.UserKeycloakRepo.LoginOtp(ctx, clientID, clientSecret, realm, username, password, totp)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (u *UserKeycloakUsecase) Logout(ctx context.Context, clientID, clientSecret, realm, refreshToken string) error {
	if err := u.UserKeycloakRepo.Logout(ctx, clientID, clientSecret, realm, refreshToken); err != nil {
		return err
	}
	return nil
}

func (u *UserKeycloakUsecase) LoginClient(ctx context.Context, clientID, clientSecret, realm string) (*keycloak.JWT, error) {
	token, err := u.UserKeycloakRepo.LoginClient(ctx, clientID, clientSecret, realm)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (u *UserKeycloakUsecase) RefreshToken(ctx context.Context, refreshToken, clientID, clientSecret, realm string) (*keycloak.JWT, error) {
	token, err := u.UserKeycloakRepo.RefreshToken(ctx, refreshToken, clientID, clientSecret, realm)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (u *UserKeycloakUsecase) GetUserInfo(ctx context.Context, accessToken, realm string) (*keycloak.UserInfo, error) {
	userInfo, err := u.UserKeycloakRepo.GetUserInfo(ctx, accessToken, realm)
	if err != nil {
		return nil, err
	}
	return userInfo, nil
}

func (u *UserKeycloakUsecase) SetPassword(ctx context.Context, token, userID, realm, password string, temporary bool) error {
	if err := u.UserKeycloakRepo.SetPassword(ctx, token, userID, realm, password, temporary); err != nil {
		return err
	}
	return nil
}

func (u *UserKeycloakUsecase) CreateGroup(ctx context.Context, accessToken, realm string, group keycloak.Group) (string, error)
func (u *UserKeycloakUsecase) UpdateUser(ctx context.Context, accessToken, realm string, user keycloak.User) error
func (u *UserKeycloakUsecase) UpdateGroup(ctx context.Context, accessToken, realm string, updatedGroup keycloak.Group) error
func (u *UserKeycloakUsecase) UpdateRole(ctx context.Context, accessToken, realm, idOfClient string, role keycloak.Role) error
func (u *UserKeycloakUsecase) UpdateClient(ctx context.Context, accessToken, realm string, updatedClient keycloak.Client) error
func (u *UserKeycloakUsecase) UpdateClientScope(ctx context.Context, accessToken, realm string, scope keycloak.ClientScope) error
func (u *UserKeycloakUsecase) DeleteUser(ctx context.Context, accessToken, realm, userID string) error
func (u *UserKeycloakUsecase) DeleteComponent(ctx context.Context, accessToken, realm, componentID string) error
func (u *UserKeycloakUsecase) DeleteGroup(ctx context.Context, accessToken, realm, groupID string) error
func (u *UserKeycloakUsecase) DeleteClientRole(ctx context.Context, accessToken, realm, idOfClient, roleName string) error
func (u *UserKeycloakUsecase) DeleteClientRoleFromUser(ctx context.Context, token, realm, idOfClient, userID string, roles []keycloak.Role) error
func (u *UserKeycloakUsecase) DeleteClient(ctx context.Context, accessToken, realm, idOfClient string) error
func (u *UserKeycloakUsecase) DeleteClientScope(ctx context.Context, accessToken, realm, scopeID string) error
func (u *UserKeycloakUsecase) DeleteClientScopeMappingsRealmRoles(ctx context.Context, token, realm, idOfClient string, roles []keycloak.Role) error
func (u *UserKeycloakUsecase) DeleteClientScopeMappingsClientRoles(ctx context.Context, token, realm, idOfClient, idOfSelectedClient string, roles []keycloak.Role) error
func (u *UserKeycloakUsecase) DeleteClientScopesScopeMappingsRealmRoles(ctx context.Context, token, realm, idOfCLientScope string, roles []keycloak.Role) error
func (u *UserKeycloakUsecase) DeleteClientScopesScopeMappingsClientRoles(ctx context.Context, token, realm, idOfClientScope, ifOfClient string, roles []keycloak.Role) error
func (u *UserKeycloakUsecase) GetClient(ctx context.Context, accessToken, realm, idOfClient string) (*keycloak.Client, error)
func (u *UserKeycloakUsecase) GetClientsDefaultScopes(ctx context.Context, token, realm, idOfClient string) ([]*keycloak.ClientScope, error)
func (u *UserKeycloakUsecase) AddDefaultScopeToClient(ctx context.Context, token, realm, idOfClient, scopeID string) error
func (u *UserKeycloakUsecase) RemoveDefaultScopeFromClient(ctx context.Context, token, realm, idOfClient, scopeID string) error
func (u *UserKeycloakUsecase) GetClientsOptionalScopes(ctx context.Context, token, realm, idOfClient string) ([]*keycloak.ClientScope, error)
func (u *UserKeycloakUsecase) AddOptionalScopeToClient(ctx context.Context, token, realm, idOfClient, scopeID string) error
func (u *UserKeycloakUsecase) RemoveOptionalScopeFromClient(ctx context.Context, token, realm, idOfClient, scopeID string) error
func (u *UserKeycloakUsecase) GetDefaultOptionalClientScopes(ctx context.Context, token, realm string) ([]*keycloak.ClientScope, error)
func (u *UserKeycloakUsecase) GetDefaultDefaultClientScopes(ctx context.Context, token, realm string) ([]*keycloak.ClientScope, error)
func (u *UserKeycloakUsecase) GetClientScope(ctx context.Context, token, realm, scopeID string) (*keycloak.ClientScope, error)
func (u *UserKeycloakUsecase) GetClientScopes(ctx context.Context, token, realm string) ([]*keycloak.ClientScope, error)
func (u *UserKeycloakUsecase) GetClientScopeMappings(ctx context.Context, token, realm, idOfClient string) (*keycloak.MappingsRepresentation, error)
func (u *UserKeycloakUsecase) GetClientScopeMappingsRealmRoles(ctx context.Context, token, realm, idOfClient string) ([]*keycloak.Role, error)
func (u *UserKeycloakUsecase) GetClientScopeMappingsRealmRolesAvailable(ctx context.Context, token, realm, idOfClient string) ([]*keycloak.Role, error)
func (u *UserKeycloakUsecase) GetClientScopesScopeMappingsRealmRolesAvailable(ctx context.Context, token, realm, idOfClientScope string) ([]*keycloak.Role, error)
func (u *UserKeycloakUsecase) GetClientScopesScopeMappingsClientRolesAvailable(ctx context.Context, token, realm, idOfClientScope, idOfClient string) ([]*keycloak.Role, error)
func (u *UserKeycloakUsecase) GetClientScopeMappingsClientRoles(ctx context.Context, token, realm, idOfClient, idOfSelectedClient string) ([]*keycloak.Role, error)
func (u *UserKeycloakUsecase) GetClientScopesScopeMappingsRealmRoles(ctx context.Context, token, realm, idOfClientScope string) ([]*keycloak.Role, error)
func (u *UserKeycloakUsecase) GetClientScopesScopeMappingsClientRoles(ctx context.Context, token, realm, idOfClientScope, idOfClient string) ([]*keycloak.Role, error)
func (u *UserKeycloakUsecase) GetClientScopeMappingsClientRolesAvailable(ctx context.Context, token, realm, idOfClient, idOfSelectedClient string) ([]*keycloak.Role, error)
func (u *UserKeycloakUsecase) GetClientSecret(ctx context.Context, token, realm, idOfClient string) (*keycloak.CredentialRepresentation, error)
func (u *UserKeycloakUsecase) GetClientServiceAccount(ctx context.Context, token, realm, idOfClient string) (*keycloak.User, error)
func (u *UserKeycloakUsecase) RegenerateClientSecret(ctx context.Context, token, realm, idOfClient string) (*keycloak.CredentialRepresentation, error)
func (u *UserKeycloakUsecase) GetKeyStoreConfig(ctx context.Context, accessToken, realm string) (*keycloak.KeyStoreConfig, error)
func (u *UserKeycloakUsecase) GetUserByID(ctx context.Context, accessToken, realm, userID string) (*keycloak.User, error)
func (u *UserKeycloakUsecase) GetUserCount(ctx context.Context, accessToken, realm string, params keycloak.GetUsersParams) (int, error)
func (u *UserKeycloakUsecase) GetUsers(ctx context.Context, accessToken, realm string, params keycloak.GetUsersParams) ([]*keycloak.User, error)
func (u *UserKeycloakUsecase) GetUserGroups(ctx context.Context, accessToken, realm, userID string, params keycloak.GetGroupsParams) ([]*keycloak.Group, error)
func (u *UserKeycloakUsecase) AddUserToGroup(ctx context.Context, token, realm, userID, groupID string) error
func (u *UserKeycloakUsecase) DeleteUserFromGroup(ctx context.Context, token, realm, userID, groupID string) error
func (u *UserKeycloakUsecase) GetComponents(ctx context.Context, accessToken, realm string) ([]*keycloak.Component, error)
func (u *UserKeycloakUsecase) GetGroups(ctx context.Context, accessToken, realm string, params keycloak.GetGroupsParams) ([]*keycloak.Group, error)
func (u *UserKeycloakUsecase) GetGroupsCount(ctx context.Context, token, realm string, params keycloak.GetGroupsParams) (int, error)
func (u *UserKeycloakUsecase) GetGroup(ctx context.Context, accessToken, realm, groupID string) (*keycloak.Group, error)
func (u *UserKeycloakUsecase) GetDefaultGroups(ctx context.Context, accessToken, realm string) ([]*keycloak.Group, error)
func (u *UserKeycloakUsecase) AddDefaultGroup(ctx context.Context, accessToken, realm, groupID string) error
func (u *UserKeycloakUsecase) RemoveDefaultGroup(ctx context.Context, accessToken, realm, groupID string) error
func (u *UserKeycloakUsecase) GetGroupMembers(ctx context.Context, accessToken, realm, groupID string, params keycloak.GetGroupsParams) ([]*keycloak.User, error)
func (u *UserKeycloakUsecase) GetRoleMappingByGroupID(ctx context.Context, accessToken, realm, groupID string) (*keycloak.MappingsRepresentation, error)
func (u *UserKeycloakUsecase) GetRoleMappingByUserID(ctx context.Context, accessToken, realm, userID string) (*keycloak.MappingsRepresentation, error)
func (u *UserKeycloakUsecase) GetClientRoles(ctx context.Context, accessToken, realm, idOfClient string, params keycloak.GetRoleParams) ([]*keycloak.Role, error)
func (u *UserKeycloakUsecase) GetClientRole(ctx context.Context, token, realm, idOfClient, roleName string) (*keycloak.Role, error)
func (u *UserKeycloakUsecase) GetClientRoleByID(ctx context.Context, accessToken, realm, roleID string) (*keycloak.Role, error)
func (u *UserKeycloakUsecase) AddClientRoleComposite(ctx context.Context, token, realm, roleID string, roles []keycloak.Role) error
func (u *UserKeycloakUsecase) DeleteClientRoleComposite(ctx context.Context, token, realm, roleID string, roles []keycloak.Role) error
func (u *UserKeycloakUsecase) GetUsersByRoleName(ctx context.Context, token, realm, roleName string, roles keycloak.GetUsersByRoleParams) ([]*keycloak.User, error)
func (u *UserKeycloakUsecase) GetUsersByClientRoleName(ctx context.Context, token, realm, idOfClient, roleName string, params keycloak.GetUsersByRoleParams) ([]*keycloak.User, error)
func (u *UserKeycloakUsecase) CreateClientProtocolMapper(ctx context.Context, token, realm, idOfClient string, mapper keycloak.ProtocolMapperRepresentation) (string, error)
func (u *UserKeycloakUsecase) UpdateClientProtocolMapper(ctx context.Context, token, realm, idOfClient, mapperID string, mapper keycloak.ProtocolMapperRepresentation) error
func (u *UserKeycloakUsecase) DeleteClientProtocolMapper(ctx context.Context, token, realm, idOfClient, mapperID string) error
func (u *UserKeycloakUsecase) GetRealm(ctx context.Context, token, realm string) (*keycloak.RealmRepresentation, error)
func (u *UserKeycloakUsecase) GetRealms(ctx context.Context, token string) ([]*keycloak.RealmRepresentation, error)
func (u *UserKeycloakUsecase) CreateRealm(ctx context.Context, token string, realm keycloak.RealmRepresentation) (string, error)
func (u *UserKeycloakUsecase) UpdateRealm(ctx context.Context, token string, realm keycloak.RealmRepresentation) error
func (u *UserKeycloakUsecase) DeleteRealm(ctx context.Context, token, realm string) error
func (u *UserKeycloakUsecase) ClearRealmCache(ctx context.Context, token, realm string) error
func (u *UserKeycloakUsecase) ClearUserCache(ctx context.Context, token, realm string) error
func (u *UserKeycloakUsecase) ClearKeysCache(ctx context.Context, token, realm string) error

func (u *UserKeycloakUsecase) GetClientUserSessions(ctx context.Context, token, realm, idOfClient string) ([]*keycloak.UserSessionRepresentation, error)
func (u *UserKeycloakUsecase) GetClientOfflineSessions(ctx context.Context, token, realm, idOfClient string) ([]*keycloak.UserSessionRepresentation, error)
func (u *UserKeycloakUsecase) GetUserSessions(ctx context.Context, token, realm, userID string) ([]*keycloak.UserSessionRepresentation, error)
func (u *UserKeycloakUsecase) GetUserOfflineSessionsForClient(ctx context.Context, token, realm, userID, idOfClient string) ([]*keycloak.UserSessionRepresentation, error)

func (u *UserKeycloakUsecase) GetResource(ctx context.Context, token, realm, idOfClient, resourceID string) (*keycloak.ResourceRepresentation, error)
func (u *UserKeycloakUsecase) GetResources(ctx context.Context, token, realm, idOfClient string, params keycloak.GetResourceParams) ([]*keycloak.ResourceRepresentation, error)
func (u *UserKeycloakUsecase) CreateResource(ctx context.Context, token, realm, idOfClient string, resource keycloak.ResourceRepresentation) (*keycloak.ResourceRepresentation, error)
func (u *UserKeycloakUsecase) UpdateResource(ctx context.Context, token, realm, idOfClient string, resource keycloak.ResourceRepresentation) error
func (u *UserKeycloakUsecase) DeleteResource(ctx context.Context, token, realm, idOfClient, resourceID string) error

func (u *UserKeycloakUsecase) GetResourceClient(ctx context.Context, token, realm, resourceID string) (*keycloak.ResourceRepresentation, error)
func (u *UserKeycloakUsecase) GetResourcesClient(ctx context.Context, token, realm string, params keycloak.GetResourceParams) ([]*keycloak.ResourceRepresentation, error)
func (u *UserKeycloakUsecase) CreateResourceClient(ctx context.Context, token, realm string, resource keycloak.ResourceRepresentation) (*keycloak.ResourceRepresentation, error)
func (u *UserKeycloakUsecase) UpdateResourceClient(ctx context.Context, token, realm string, resource keycloak.ResourceRepresentation) error
func (u *UserKeycloakUsecase) DeleteResourceClient(ctx context.Context, token, realm, resourceID string) error

func (u *UserKeycloakUsecase) GetScope(ctx context.Context, token, realm, idOfClient, scopeID string) (*keycloak.ScopeRepresentation, error)
func (u *UserKeycloakUsecase) GetScopes(ctx context.Context, token, realm, idOfClient string, params keycloak.GetScopeParams) ([]*keycloak.ScopeRepresentation, error)
func (u *UserKeycloakUsecase) CreateScope(ctx context.Context, token, realm, idOfClient string, scope keycloak.ScopeRepresentation) (*keycloak.ScopeRepresentation, error)
func (u *UserKeycloakUsecase) UpdateScope(ctx context.Context, token, realm, idOfClient string, resource keycloak.ScopeRepresentation) error
func (u *UserKeycloakUsecase) DeleteScope(ctx context.Context, token, realm, idOfClient, scopeID string) error

func (u *UserKeycloakUsecase) GetPolicy(ctx context.Context, token, realm, idOfClient, policyID string) (*keycloak.PolicyRepresentation, error)
func (u *UserKeycloakUsecase) GetPolicies(ctx context.Context, token, realm, idOfClient string, params keycloak.GetPolicyParams) ([]*keycloak.PolicyRepresentation, error)
func (u *UserKeycloakUsecase) CreatePolicy(ctx context.Context, token, realm, idOfClient string, policy keycloak.PolicyRepresentation) (*keycloak.PolicyRepresentation, error)
func (u *UserKeycloakUsecase) UpdatePolicy(ctx context.Context, token, realm, idOfClient string, policy keycloak.PolicyRepresentation) error
func (u *UserKeycloakUsecase) DeletePolicy(ctx context.Context, token, realm, idOfClient, policyID string) error

func (u *UserKeycloakUsecase) GetResourcePolicy(ctx context.Context, token, realm, permissionID string) (*keycloak.ResourcePolicyRepresentation, error)
func (u *UserKeycloakUsecase) GetResourcePolicies(ctx context.Context, token, realm string, params keycloak.GetResourcePoliciesParams) ([]*keycloak.ResourcePolicyRepresentation, error)
func (u *UserKeycloakUsecase) CreateResourcePolicy(ctx context.Context, token, realm, resourceID string, policy keycloak.ResourcePolicyRepresentation) (*keycloak.ResourcePolicyRepresentation, error)
func (u *UserKeycloakUsecase) UpdateResourcePolicy(ctx context.Context, token, realm, permissionID string, policy keycloak.ResourcePolicyRepresentation) error
func (u *UserKeycloakUsecase) DeleteResourcePolicy(ctx context.Context, token, realm, permissionID string) error

func (u *UserKeycloakUsecase) GetPermission(ctx context.Context, token, realm, idOfClient, permissionID string) (*keycloak.PermissionRepresentation, error)
func (u *UserKeycloakUsecase) GetPermissions(ctx context.Context, token, realm, idOfClient string, params keycloak.GetPermissionParams) ([]*keycloak.PermissionRepresentation, error)
func (u *UserKeycloakUsecase) GetPermissionResources(ctx context.Context, token, realm, idOfClient, permissionID string) ([]*keycloak.PermissionResource, error)
func (u *UserKeycloakUsecase) GetPermissionScopes(ctx context.Context, token, realm, idOfClient, permissionID string) ([]*keycloak.PermissionScope, error)
func (u *UserKeycloakUsecase) GetDependentPermissions(ctx context.Context, token, realm, idOfClient, policyID string) ([]*keycloak.PermissionRepresentation, error)
func (u *UserKeycloakUsecase) CreatePermission(ctx context.Context, token, realm, idOfClient string, permission keycloak.PermissionRepresentation) (*keycloak.PermissionRepresentation, error)
func (u *UserKeycloakUsecase) UpdatePermission(ctx context.Context, token, realm, idOfClient string, permission keycloak.PermissionRepresentation) error
func (u *UserKeycloakUsecase) DeletePermission(ctx context.Context, token, realm, idOfClient, permissionID string) error

func (u *UserKeycloakUsecase) CreatePermissionTicket(ctx context.Context, token, realm string, permissions []keycloak.CreatePermissionTicketParams) (*keycloak.PermissionTicketResponseRepresentation, error)
func (u *UserKeycloakUsecase) GrantUserPermission(ctx context.Context, token, realm string, permission keycloak.PermissionGrantParams) (*keycloak.PermissionGrantResponseRepresentation, error)
func (u *UserKeycloakUsecase) UpdateUserPermission(ctx context.Context, token, realm string, permission keycloak.PermissionGrantParams) (*keycloak.PermissionGrantResponseRepresentation, error)
func (u *UserKeycloakUsecase) GetUserPermissions(ctx context.Context, token, realm string, params keycloak.GetUserPermissionParams) ([]*keycloak.PermissionGrantResponseRepresentation, error)
func (u *UserKeycloakUsecase) DeleteUserPermission(ctx context.Context, token, realm, ticketID string) error

func (u *UserKeycloakUsecase) GetCredentialRegistrators(ctx context.Context, token, realm string) ([]string, error)
func (u *UserKeycloakUsecase) GetConfiguredUserStorageCredentialTypes(ctx context.Context, token, realm, userID string) ([]string, error)
func (u *UserKeycloakUsecase) GetCredentials(ctx context.Context, token, realm, UserID string) ([]*keycloak.CredentialRepresentation, error)
func (u *UserKeycloakUsecase) DeleteCredentials(ctx context.Context, token, realm, UserID, CredentialID string) error
func (u *UserKeycloakUsecase) UpdateCredentialUserLabel(ctx context.Context, token, realm, userID, credentialID, userLabel string) error
func (u *UserKeycloakUsecase) DisableAllCredentialsByType(ctx context.Context, token, realm, userID string, types []string) error
func (u *UserKeycloakUsecase) MoveCredentialBehind(ctx context.Context, token, realm, userID, credentialID, newPreviousCredentialID string) error
func (u *UserKeycloakUsecase) MoveCredentialToFirst(ctx context.Context, token, realm, userID, credentialID string) error

func (u *UserKeycloakUsecase) GetAuthenticationFlows(ctx context.Context, token, realm string) ([]*keycloak.AuthenticationFlowRepresentation, error)
func (u *UserKeycloakUsecase) GetAuthenticationFlow(ctx context.Context, token, realm string, authenticationFlowID string) (*keycloak.AuthenticationFlowRepresentation, error)
func (u *UserKeycloakUsecase) CreateAuthenticationFlow(ctx context.Context, token, realm string, flow keycloak.AuthenticationFlowRepresentation) error
func (u *UserKeycloakUsecase) UpdateAuthenticationFlow(ctx context.Context, token, realm string, flow keycloak.AuthenticationFlowRepresentation, authenticationFlowID string) (*keycloak.AuthenticationFlowRepresentation, error)
func (u *UserKeycloakUsecase) DeleteAuthenticationFlow(ctx context.Context, token, realm, flowID string) error

func (u *UserKeycloakUsecase) CreateIdentityProvider(ctx context.Context, token, realm string, providerRep keycloak.IdentityProviderRepresentation) (string, error)
func (u *UserKeycloakUsecase) GetIdentityProvider(ctx context.Context, token, realm, alias string) (*keycloak.IdentityProviderRepresentation, error)
func (u *UserKeycloakUsecase) GetIdentityProviders(ctx context.Context, token, realm string) ([]*keycloak.IdentityProviderRepresentation, error)
func (u *UserKeycloakUsecase) UpdateIdentityProvider(ctx context.Context, token, realm, alias string, providerRep keycloak.IdentityProviderRepresentation) error
func (u *UserKeycloakUsecase) DeleteIdentityProvider(ctx context.Context, token, realm, alias string) error

func (u *UserKeycloakUsecase) CreateIdentityProviderMapper(ctx context.Context, token, realm, alias string, mapper keycloak.IdentityProviderMapper) (string, error)
func (u *UserKeycloakUsecase) GetIdentityProviderMapper(ctx context.Context, token string, realm string, alias string, mapperID string) (*keycloak.IdentityProviderMapper, error)
func (u *UserKeycloakUsecase) CreateUserFederatedIdentity(ctx context.Context, token, realm, userID, providerID string, federatedIdentityRep keycloak.FederatedIdentityRepresentation) error
func (u *UserKeycloakUsecase) GetUserFederatedIdentities(ctx context.Context, token, realm, userID string) ([]*keycloak.FederatedIdentityRepresentation, error)
func (u *UserKeycloakUsecase) DeleteUserFederatedIdentity(ctx context.Context, token, realm, userID, providerID string) error

func (u *UserKeycloakUsecase) GetEvents(ctx context.Context, token string, realm string, params keycloak.GetEventsParams) ([]*keycloak.EventRepresentation, error)
