package keycloak

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/manage-service/internal/domain/model"
	"github.com/JIeeiroSst/manage-service/internal/domain/port"
	"github.com/JIeeiroSst/manage-service/pkg/log"

	"github.com/Nerzal/gocloak/v13"
)

type IdentityProvider struct {
	client *gocloak.GoCloak
}

func NewIdentityProvider(client *gocloak.GoCloak) port.IdentityProvider {
	return &IdentityProvider{
		client: client,
	}
}

func (r *IdentityProvider) LoginAdmin(ctx context.Context, user model.LoginAdmin) (*model.Token, error) {
	token, err := r.client.LoginAdmin(ctx, user.User, user.Password, user.RealmName)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return &model.Token{
		Token: token.AccessToken,
	}, nil
}

func (r *IdentityProvider) CreateUser(ctx context.Context, user model.CreateUser) error {
	userKeycloak := gocloak.User{
		FirstName: gocloak.StringP(user.FirstName),
		LastName:  gocloak.StringP(user.LastName),
		Email:     gocloak.StringP(user.Email),
		Enabled:   gocloak.BoolP(user.Enabled),
		Username:  gocloak.StringP(user.Username),
	}
	_, err := r.client.CreateUser(ctx, user.Token, user.Realm, userKeycloak)
	if err != nil {
		log.Error(err)
		return err

	}
	return nil
}

func (r *IdentityProvider) IntrospectToken(ctx context.Context, token model.IntrospectToken) (*[]gocloak.ResourcePermission, error) {
	rptResult, err := r.client.RetrospectToken(ctx, token.Token, token.ClientID, token.ClientSecret, token.Realm)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	if rptResult.Active == nil || !*rptResult.Active {
		log.Error("Token is not active")
		return nil, errors.New("Token is not active")
	}
	permissions := rptResult.Permissions
	return permissions, nil
}

func (r *IdentityProvider) GetTokenUser(ctx context.Context, realm string) (*model.TokenInfo, error) {
	options := gocloak.TokenOptions{}
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

func (r *IdentityProvider) GetClients(ctx context.Context, user model.Client) ([]*gocloak.Client, error) {
	clients, err := r.client.GetClients(
		ctx,
		user.Token,
		user.Realm,
		gocloak.GetClientsParams{
			ClientID: &user.ClientName,
		},
	)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return clients, nil
}

func (r *IdentityProvider) Login(ctx context.Context, clientID, clientSecret, realm, username, password string) (*gocloak.JWT, error) {
	token, err := r.client.Login(ctx, clientID, clientSecret, realm, username, password)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (r *IdentityProvider) LoginOtp(ctx context.Context, clientID, clientSecret, realm, username, password, totp string) (*gocloak.JWT, error) {
	token, err := r.client.LoginOtp(ctx, clientID, clientSecret, realm, username, password, totp)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (r *IdentityProvider) Logout(ctx context.Context, clientID, clientSecret, realm, refreshToken string) error {
	if err := r.client.Logout(ctx, clientID, clientSecret, realm, refreshToken); err != nil {
		return err
	}
	return nil
}

func (r *IdentityProvider) LoginClient(ctx context.Context, clientID, clientSecret, realm string) (*gocloak.JWT, error) {
	token, err := r.client.LoginClient(ctx, clientID, clientSecret, realm)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (r *IdentityProvider) RefreshToken(ctx context.Context, refreshToken, clientID, clientSecret, realm string) (*gocloak.JWT, error) {
	token, err := r.client.RefreshToken(ctx, refreshToken, clientID, clientSecret, realm)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (r *IdentityProvider) GetUserInfo(ctx context.Context, accessToken, realm string) (*gocloak.UserInfo, error) {
	userInfo, err := r.client.GetUserInfo(ctx, accessToken, realm)
	if err != nil {
		return nil, err
	}
	return userInfo, nil
}

func (r *IdentityProvider) SetPassword(ctx context.Context, token, userID, realm, password string, temporary bool) error {
	if err := r.client.SetPassword(ctx, token, userID, realm, password, temporary); err != nil {
		return err
	}
	return nil
}

func (r *IdentityProvider) CreateGroup(ctx context.Context, accessToken, realm string, group gocloak.Group) (string, error) {
	msg, err := r.client.CreateGroup(ctx, accessToken, realm, group)
	if err != nil {
		return "", err
	}
	return msg, nil
}

func (r *IdentityProvider) UpdateUser(ctx context.Context, accessToken, realm string, user gocloak.User) error {
	if err := r.client.UpdateUser(ctx, accessToken, realm, user); err != nil {
		return err
	}
	return nil
}

func (r *IdentityProvider) UpdateGroup(ctx context.Context, accessToken, realm string, updatedGroup gocloak.Group) error {
	if err := r.client.UpdateGroup(ctx, accessToken, realm, updatedGroup); err != nil {
		return err
	}
	return nil
}

func (r *IdentityProvider) UpdateRole(ctx context.Context, accessToken, realm, idOfClient string, role gocloak.Role) error {
	if err := r.client.UpdateRole(ctx, accessToken, realm, idOfClient, role); err != nil {
		return err
	}
	return nil
}

func (r *IdentityProvider) UpdateClient(ctx context.Context, accessToken, realm string, updatedClient gocloak.Client) error {
	if err := r.client.UpdateClient(ctx, accessToken, realm, updatedClient); err != nil {
		return err
	}
	return nil
}

func (r *IdentityProvider) UpdateClientScope(ctx context.Context, accessToken, realm string, scope gocloak.ClientScope) error {
	if err := r.client.UpdateClientScope(ctx, accessToken, realm, scope); err != nil {
		return err
	}
	return nil
}

func (r *IdentityProvider) DeleteUser(ctx context.Context, accessToken, realm, userID string) error {
	if err := r.client.DeleteUser(ctx, accessToken, realm, userID); err != nil {
		return err
	}
	return nil
}

func (r *IdentityProvider) DeleteComponent(ctx context.Context, accessToken, realm, componentID string) error {
	if err := r.client.DeleteComponent(ctx, accessToken, realm, componentID); err != nil {
		return err
	}
	return nil
}

func (r *IdentityProvider) DeleteGroup(ctx context.Context, accessToken, realm, groupID string) error {
	if err := r.client.DeleteGroup(ctx, accessToken, realm, groupID); err != nil {
		return err
	}
	return nil
}

func (r *IdentityProvider) DeleteClientRole(ctx context.Context, accessToken, realm, idOfClient, roleName string) error {
	if err := r.client.DeleteClientRole(ctx, accessToken, realm, idOfClient, roleName); err != nil {
		return err
	}
	return nil
}

func (r *IdentityProvider) DeleteClientRoleFromUser(ctx context.Context, token, realm, idOfClient, userID string, roles []gocloak.Role) error {
	if err := r.client.DeleteClientRoleFromUser(ctx, token, realm, idOfClient, userID, roles); err != nil {
		return err
	}
	return nil
}

func (r *IdentityProvider) DeleteClient(ctx context.Context, accessToken, realm, idOfClient string) error {
	if err := r.client.DeleteClient(ctx, accessToken, realm, idOfClient); err != nil {
		return err
	}
	return nil
}

func (r *IdentityProvider) DeleteClientScope(ctx context.Context, accessToken, realm, scopeID string) error {
	if err := r.client.DeleteClientScope(ctx, accessToken, realm, scopeID); err != nil {
		return err
	}
	return nil
}

func (r *IdentityProvider) DeleteClientScopeMappingsRealmRoles(ctx context.Context, token, realm, idOfClient string, roles []gocloak.Role) error {
	if err := r.client.DeleteClientScopeMappingsRealmRoles(ctx, token, realm, idOfClient, roles); err != nil {
		return err
	}
	return nil
}

func (r *IdentityProvider) DeleteClientScopeMappingsClientRoles(ctx context.Context, token, realm, idOfClient, idOfSelectedClient string, roles []gocloak.Role) error {
	if err := r.client.DeleteClientScopeMappingsClientRoles(ctx, token, realm, idOfClient, idOfSelectedClient, roles); err != nil {
		return err
	}
	return nil
}

func (r *IdentityProvider) DeleteClientScopesScopeMappingsRealmRoles(ctx context.Context, token, realm, idOfCLientScope string, roles []gocloak.Role) error {
	if err := r.client.DeleteClientScopesScopeMappingsRealmRoles(ctx, token, realm, idOfCLientScope, roles); err != nil {
		return err
	}
	return nil
}

func (r *IdentityProvider) DeleteClientScopesScopeMappingsClientRoles(ctx context.Context, token, realm, idOfClientScope, ifOfClient string, roles []gocloak.Role) error {
	if err := r.client.DeleteClientScopesScopeMappingsClientRoles(ctx, token, realm, idOfClientScope, ifOfClient, roles); err != nil {
		return err
	}
	return nil
}

func (r *IdentityProvider) GetClient(ctx context.Context, accessToken, realm, idOfClient string) (*gocloak.Client, error) {
	client, err := r.client.GetClient(ctx, accessToken, realm, idOfClient)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (r *IdentityProvider) GetClientsDefaultScopes(ctx context.Context, token, realm, idOfClient string) ([]*gocloak.ClientScope, error) {
	clientScope, err := r.client.GetClientsDefaultScopes(ctx, token, realm, idOfClient)
	if err != nil {
		return nil, err
	}
	return clientScope, nil
}

func (r *IdentityProvider) AddDefaultScopeToClient(ctx context.Context, token, realm, idOfClient, scopeID string) error {
	if err := r.client.AddDefaultScopeToClient(ctx, token, realm, idOfClient, scopeID); err != nil {
		return err
	}
	return nil
}

func (r *IdentityProvider) RemoveDefaultScopeFromClient(ctx context.Context, token, realm, idOfClient, scopeID string) error {
	if err := r.client.RemoveDefaultScopeFromClient(ctx, token, realm, idOfClient, scopeID); err != nil {
		return err
	}
	return nil
}

func (r *IdentityProvider) GetClientsOptionalScopes(ctx context.Context, token, realm, idOfClient string) ([]*gocloak.ClientScope, error) {
	clientScope, err := r.client.GetClientsOptionalScopes(ctx, token, realm, idOfClient)
	if err != nil {
		return nil, err
	}
	return clientScope, nil
}

func (r *IdentityProvider) AddOptionalScopeToClient(ctx context.Context, token, realm, idOfClient, scopeID string) error {
	if err := r.client.AddOptionalScopeToClient(ctx, token, realm, idOfClient, scopeID); err != nil {
		return err
	}
	return nil
}

func (r *IdentityProvider) RemoveOptionalScopeFromClient(ctx context.Context, token, realm, idOfClient, scopeID string) error {
	if err := r.client.RemoveOptionalScopeFromClient(ctx, token, realm, idOfClient, scopeID); err != nil {
		return err
	}
	return nil
}

func (r *IdentityProvider) GetDefaultOptionalClientScopes(ctx context.Context, token, realm string) ([]*gocloak.ClientScope, error) {
	clientScope, err := r.client.GetDefaultOptionalClientScopes(ctx, token, realm)
	if err != nil {
		return nil, err
	}
	return clientScope, nil
}

func (r *IdentityProvider) GetDefaultDefaultClientScopes(ctx context.Context, token, realm string) ([]*gocloak.ClientScope, error) {
	clientScope, err := r.client.GetDefaultDefaultClientScopes(ctx, token, realm)
	if err != nil {
		return nil, err
	}
	return clientScope, nil
}

func (r *IdentityProvider) GetClientScope(ctx context.Context, token, realm, scopeID string) (*gocloak.ClientScope, error) {
	clientScope, err := r.client.GetClientScope(ctx, token, realm, scopeID)
	if err != nil {
		return nil, err
	}
	return clientScope, nil
}

func (r *IdentityProvider) GetClientScopes(ctx context.Context, token, realm string) ([]*gocloak.ClientScope, error) {
	clientScope, err := r.client.GetClientScopes(ctx, token, realm)
	if err != nil {
		return nil, err
	}
	return clientScope, nil
}

func (r *IdentityProvider) GetClientScopeMappings(ctx context.Context, token, realm, idOfClient string) (*gocloak.MappingsRepresentation, error) {
	mappingsRepresentation, err := r.client.GetClientScopeMappings(ctx, token, realm, idOfClient)
	if err != nil {
		return nil, err
	}
	return mappingsRepresentation, nil
}

func (r *IdentityProvider) GetClientScopeMappingsRealmRoles(ctx context.Context, token, realm, idOfClient string) ([]*gocloak.Role, error) {
	role, err := r.client.GetClientScopeMappingsRealmRoles(ctx, token, realm, idOfClient)
	if err != nil {
		return nil, err
	}
	return role, nil
}

func (r *IdentityProvider) GetClientScopeMappingsRealmRolesAvailable(ctx context.Context, token, realm, idOfClient string) ([]*gocloak.Role, error) {
	role, err := r.client.GetClientScopeMappingsRealmRolesAvailable(ctx, token, realm, idOfClient)
	if err != nil {
		return nil, err
	}
	return role, nil
}

func (r *IdentityProvider) GetClientScopesScopeMappingsRealmRolesAvailable(ctx context.Context, token, realm, idOfClientScope string) ([]*gocloak.Role, error) {
	role, err := r.client.GetClientScopesScopeMappingsRealmRolesAvailable(ctx, token, realm, idOfClientScope)
	if err != nil {
		return nil, err
	}
	return role, nil
}

func (r *IdentityProvider) GetClientScopesScopeMappingsClientRolesAvailable(ctx context.Context, token, realm, idOfClientScope, idOfClient string) ([]*gocloak.Role, error) {
	role, err := r.client.GetClientScopesScopeMappingsClientRolesAvailable(ctx, token, realm, idOfClientScope, idOfClient)
	if err != nil {
		return nil, err
	}
	return role, nil
}

func (r *IdentityProvider) GetClientScopeMappingsClientRoles(ctx context.Context, token, realm, idOfClient, idOfSelectedClient string) ([]*gocloak.Role, error) {
	role, err := r.client.GetClientScopeMappingsClientRoles(ctx, token, realm, idOfClient, idOfSelectedClient)
	if err != nil {
		return nil, err
	}
	return role, nil
}

func (r *IdentityProvider) GetClientScopesScopeMappingsRealmRoles(ctx context.Context, token, realm, idOfClientScope string) ([]*gocloak.Role, error) {
	role, err := r.client.GetClientScopesScopeMappingsRealmRoles(ctx, token, realm, idOfClientScope)
	if err != nil {
		return nil, err
	}
	return role, nil
}

func (r *IdentityProvider) GetClientScopesScopeMappingsClientRoles(ctx context.Context, token, realm, idOfClientScope, idOfClient string) ([]*gocloak.Role, error) {
	role, err := r.client.GetClientScopesScopeMappingsClientRoles(ctx, token, realm, idOfClientScope, idOfClient)
	if err != nil {
		return nil, err
	}
	return role, nil
}

func (r *IdentityProvider) GetClientScopeMappingsClientRolesAvailable(ctx context.Context, token, realm, idOfClient, idOfSelectedClient string) ([]*gocloak.Role, error) {
	role, err := r.client.GetClientScopeMappingsClientRolesAvailable(ctx, token, realm, idOfClient, idOfSelectedClient)
	if err != nil {
		return nil, err
	}
	return role, nil
}

func (r *IdentityProvider) GetClientSecret(ctx context.Context, token, realm, idOfClient string) (*gocloak.CredentialRepresentation, error) {
	clientSecret, err := r.client.GetClientSecret(ctx, token, realm, idOfClient)
	if err != nil {
		return nil, err
	}
	return clientSecret, nil
}

func (r *IdentityProvider) GetClientServiceAccount(ctx context.Context, token, realm, idOfClient string) (*gocloak.User, error) {
	user, err := r.client.GetClientServiceAccount(ctx, token, realm, idOfClient)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *IdentityProvider) RegenerateClientSecret(ctx context.Context, token, realm, idOfClient string) (*gocloak.CredentialRepresentation, error) {
	credentialRepresentation, err := r.client.RegenerateClientSecret(ctx, token, realm, idOfClient)
	if err != nil {
		return nil, err
	}
	return credentialRepresentation, nil
}

func (r *IdentityProvider) GetKeyStoreConfig(ctx context.Context, accessToken, realm string) (*gocloak.KeyStoreConfig, error) {
	keyconfig, err := r.client.GetKeyStoreConfig(ctx, accessToken, realm)
	if err != nil {
		return nil, err
	}
	return keyconfig, nil
}

func (r *IdentityProvider) GetUserByID(ctx context.Context, accessToken, realm, userID string) (*gocloak.User, error) {
	user, err := r.client.GetUserByID(ctx, accessToken, realm, userID)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *IdentityProvider) GetUserCount(ctx context.Context, accessToken, realm string, params gocloak.GetUsersParams) (int, error) {
	count, err := r.client.GetUserCount(ctx, accessToken, realm, params)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *IdentityProvider) GetUsers(ctx context.Context, accessToken, realm string, params gocloak.GetUsersParams) ([]*gocloak.User, error) {
	user, err := r.client.GetUsers(ctx, accessToken, realm, params)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *IdentityProvider) GetUserGroups(ctx context.Context, accessToken, realm, userID string, params gocloak.GetGroupsParams) ([]*gocloak.Group, error) {
	userGroup, err := r.client.GetUserGroups(ctx, accessToken, realm, userID, params)
	if err != nil {
		return nil, err
	}
	return userGroup, nil
}

func (r *IdentityProvider) AddUserToGroup(ctx context.Context, token, realm, userID, groupID string) error {
	if err := r.client.AddUserToGroup(ctx, token, realm, userID, groupID); err != nil {
		return err
	}
	return nil
}

func (r *IdentityProvider) DeleteUserFromGroup(ctx context.Context, token, realm, userID, groupID string) error {
	if err := r.client.DeleteUserFromGroup(ctx, token, realm, userID, groupID); err != nil {
		return err
	}
	return nil
}

func (r *IdentityProvider) GetComponents(ctx context.Context, accessToken, realm string) ([]*gocloak.Component, error) {
	component, err := r.client.GetComponents(ctx, accessToken, realm)
	if err != nil {
		return nil, err
	}
	return component, nil
}

func (r *IdentityProvider) GetGroups(ctx context.Context, accessToken, realm string, params gocloak.GetGroupsParams) ([]*gocloak.Group, error) {
	group, err := r.client.GetGroups(ctx, accessToken, realm, params)
	if err != nil {
		return nil, err
	}
	return group, nil
}

func (r *IdentityProvider) GetGroupsCount(ctx context.Context, token, realm string, params gocloak.GetGroupsParams) (int, error) {
	count, err := r.client.GetGroupsCount(ctx, token, realm, params)
	if err != nil {
		return 0, nil
	}
	return count, nil
}

func (r *IdentityProvider) GetGroup(ctx context.Context, accessToken, realm, groupID string) (*gocloak.Group, error) {
	group, err := r.client.GetGroup(ctx, accessToken, realm, groupID)
	if err != nil {
		return nil, err
	}
	return group, nil
}

func (r *IdentityProvider) GetDefaultGroups(ctx context.Context, accessToken, realm string) ([]*gocloak.Group, error) {
	group, err := r.client.GetDefaultGroups(ctx, accessToken, realm)
	if err != nil {
		return nil, err
	}
	return group, nil
}

func (r *IdentityProvider) AddDefaultGroup(ctx context.Context, accessToken, realm, groupID string) error {
	if err := r.client.AddDefaultGroup(ctx, accessToken, realm, groupID); err != nil {
		return err
	}
	return nil
}

func (r *IdentityProvider) RemoveDefaultGroup(ctx context.Context, accessToken, realm, groupID string) error {
	if err := r.client.RemoveDefaultGroup(ctx, accessToken, realm, groupID); err != nil {
		return err
	}
	return nil
}

func (r *IdentityProvider) GetGroupMembers(ctx context.Context, accessToken, realm, groupID string, params gocloak.GetGroupsParams) ([]*gocloak.User, error) {
	user, err := r.client.GetGroupMembers(ctx, accessToken, realm, groupID, params)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *IdentityProvider) GetRoleMappingByGroupID(ctx context.Context, accessToken, realm, groupID string) (*gocloak.MappingsRepresentation, error) {
	mappingsRepresentation, err := r.client.GetRoleMappingByGroupID(ctx, accessToken, realm, groupID)
	if err != nil {
		return nil, err
	}
	return mappingsRepresentation, nil
}

func (r *IdentityProvider) GetRoleMappingByUserID(ctx context.Context, accessToken, realm, userID string) (*gocloak.MappingsRepresentation, error) {
	mappingsRepresentation, err := r.client.GetRoleMappingByUserID(ctx, accessToken, realm, userID)
	if err != nil {
		return nil, err
	}
	return mappingsRepresentation, nil
}

func (r *IdentityProvider) GetClientRoles(ctx context.Context, accessToken, realm, idOfClient string, params gocloak.GetRoleParams) ([]*gocloak.Role, error) {
	role, err := r.client.GetClientRoles(ctx, accessToken, realm, idOfClient, params)
	if err != nil {
		return nil, err
	}
	return role, nil
}

func (r *IdentityProvider) GetClientRole(ctx context.Context, token, realm, idOfClient, roleName string) (*gocloak.Role, error) {
	role, err := r.client.GetClientRole(ctx, token, realm, idOfClient, roleName)
	if err != nil {
		return nil, err
	}
	return role, nil
}

func (r *IdentityProvider) GetClientRoleByID(ctx context.Context, accessToken, realm, roleID string) (*gocloak.Role, error) {
	role, err := r.client.GetClientRoleByID(ctx, accessToken, realm, roleID)
	if err != nil {
		return nil, err
	}
	return role, nil
}

func (r *IdentityProvider) AddClientRoleComposite(ctx context.Context, token, realm, roleID string, roles []gocloak.Role) error {
	if err := r.client.AddClientRoleComposite(ctx, token, realm, roleID, roles); err != nil {
		return err
	}
	return nil
}

func (r *IdentityProvider) DeleteClientRoleComposite(ctx context.Context, token, realm, roleID string, roles []gocloak.Role) error {
	if err := r.client.DeleteClientRoleComposite(ctx, token, realm, roleID, roles); err != nil {
		return err
	}
	return nil
}

func (r *IdentityProvider) GetUsersByRoleName(ctx context.Context, token, realm, roleName string, roles gocloak.GetUsersByRoleParams) ([]*gocloak.User, error) {
	user, err := r.client.GetUsersByRoleName(ctx, token, realm, roleName, roles)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *IdentityProvider) GetUsersByClientRoleName(ctx context.Context, token, realm, idOfClient, roleName string, params gocloak.GetUsersByRoleParams) ([]*gocloak.User, error) {
	user, err := r.client.GetUsersByClientRoleName(ctx, token, realm, idOfClient, roleName, params)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *IdentityProvider) CreateClientProtocolMapper(ctx context.Context, token, realm, idOfClient string, mapper gocloak.ProtocolMapperRepresentation) (string, error) {
	proto, err := r.client.CreateClientProtocolMapper(ctx, token, realm, idOfClient, mapper)
	if err != nil {
		return "", err
	}
	return proto, nil
}

func (r *IdentityProvider) UpdateClientProtocolMapper(ctx context.Context, token, realm, idOfClient, mapperID string, mapper gocloak.ProtocolMapperRepresentation) error {
	if err := r.client.UpdateClientProtocolMapper(ctx, token, realm, idOfClient, mapperID, mapper); err != nil {
		return err
	}
	return nil
}

func (r *IdentityProvider) DeleteClientProtocolMapper(ctx context.Context, token, realm, idOfClient, mapperID string) error {
	if err := r.client.DeleteClientProtocolMapper(ctx, token, realm, idOfClient, mapperID); err != nil {
		return err
	}
	return nil
}

func (r *IdentityProvider) GetRealm(ctx context.Context, token, realm string) (*gocloak.RealmRepresentation, error) {
	representation, err := r.client.GetRealm(ctx, token, realm)
	if err != nil {
		return nil, err
	}
	return representation, nil
}

func (r *IdentityProvider) GetRealms(ctx context.Context, token string) ([]*gocloak.RealmRepresentation, error) {
	representation, err := r.client.GetRealms(ctx, token)
	if err != nil {
		return nil, err
	}
	return representation, err
}

func (r *IdentityProvider) CreateRealm(ctx context.Context, token string, realm gocloak.RealmRepresentation) (string, error) {
	return r.client.CreateRealm(ctx, token, realm)
}

func (r *IdentityProvider) UpdateRealm(ctx context.Context, token string, realm gocloak.RealmRepresentation) error {
	if err := r.client.UpdateRealm(ctx, token, realm); err != nil {
		return err
	}
	return nil
}

func (r *IdentityProvider) DeleteRealm(ctx context.Context, token, realm string) error {
	if err := r.client.DeleteRealm(ctx, token, realm); err != nil {
		return err
	}
	return nil
}

func (r *IdentityProvider) ClearRealmCache(ctx context.Context, token, realm string) error {
	if err := r.client.ClearRealmCache(ctx, token, realm); err != nil {
		return err
	}
	return nil
}

func (r *IdentityProvider) ClearUserCache(ctx context.Context, token, realm string) error {
	if err := r.client.ClearUserCache(ctx, token, realm); err != nil {
		return err
	}
	return nil
}

func (r *IdentityProvider) ClearKeysCache(ctx context.Context, token, realm string) error {
	if err := r.client.ClearKeysCache(ctx, token, realm); err != nil {
		return err
	}
	return nil
}

func (r *IdentityProvider) GetClientUserSessions(ctx context.Context, token, realm, idOfClient string) ([]*gocloak.UserSessionRepresentation, error) {
	userSessionRepresentation, err := r.client.GetClientUserSessions(ctx, token, realm, idOfClient)
	if err != nil {
		return nil, err
	}
	return userSessionRepresentation, nil
}
func (r *IdentityProvider) GetClientOfflineSessions(ctx context.Context, token, realm, idOfClient string) ([]*gocloak.UserSessionRepresentation, error) {
	userSessionRepresentation, err := r.client.GetClientOfflineSessions(ctx, token, realm, idOfClient)
	if err != nil {
		return nil, err
	}
	return userSessionRepresentation, nil
}

func (r *IdentityProvider) GetUserSessions(ctx context.Context, token, realm, userID string) ([]*gocloak.UserSessionRepresentation, error) {
	userSessionRepresentation, err := r.client.GetUserSessions(ctx, token, realm, userID)
	if err != nil {
		return nil, err
	}
	return userSessionRepresentation, nil
}

func (r *IdentityProvider) GetUserOfflineSessionsForClient(ctx context.Context, token, realm, userID, idOfClient string) ([]*gocloak.UserSessionRepresentation, error) {
	userSessionRepresentation, err := r.client.GetUserOfflineSessionsForClient(ctx, token, realm, userID, idOfClient)
	if err != nil {
		return nil, err
	}
	return userSessionRepresentation, nil
}

func (r *IdentityProvider) GetResource(ctx context.Context, token, realm, idOfClient, resourceID string) (*gocloak.ResourceRepresentation, error) {
	resourceRepresentation, err := r.client.GetResource(ctx, token, realm, idOfClient, resourceID)
	if err != nil {
		return nil, err
	}
	return resourceRepresentation, nil
}

func (r *IdentityProvider) GetResources(ctx context.Context, token, realm, idOfClient string, params gocloak.GetResourceParams) ([]*gocloak.ResourceRepresentation, error) {
	resourceRepresentation, err := r.client.GetResources(ctx, token, realm, idOfClient, params)
	if err != nil {
		return nil, err
	}
	return resourceRepresentation, nil
}

func (r *IdentityProvider) CreateResource(ctx context.Context, token, realm, idOfClient string, resource gocloak.ResourceRepresentation) (*gocloak.ResourceRepresentation, error) {
	resourceRepresentation, err := r.client.CreateResource(ctx, token, realm, idOfClient, resource)
	if err != nil {
		return nil, err
	}
	return resourceRepresentation, nil
}

func (r *IdentityProvider) UpdateResource(ctx context.Context, token, realm, idOfClient string, resource gocloak.ResourceRepresentation) error {
	if err := r.client.UpdateResource(ctx, token, realm, idOfClient, resource); err != nil {
		return err
	}
	return nil
}

func (r *IdentityProvider) DeleteResource(ctx context.Context, token, realm, idOfClient, resourceID string) error {
	if err := r.client.DeleteResource(ctx, token, realm, idOfClient, resourceID); err != nil {
		return err
	}
	return nil
}

func (r *IdentityProvider) GetResourceClient(ctx context.Context, token, realm, resourceID string) (*gocloak.ResourceRepresentation, error) {
	resourceRepresentation, err := r.client.GetResourceClient(ctx, token, realm, resourceID)
	if err != nil {
		return nil, err
	}
	return resourceRepresentation, nil
}

func (r *IdentityProvider) GetResourcesClient(ctx context.Context, token, realm string, params gocloak.GetResourceParams) ([]*gocloak.ResourceRepresentation, error) {
	resourceRepresentation, err := r.client.GetResourcesClient(ctx, token, realm, params)
	if err != nil {
		return nil, err
	}
	return resourceRepresentation, nil
}

func (r *IdentityProvider) CreateResourceClient(ctx context.Context, token, realm string, resource gocloak.ResourceRepresentation) (*gocloak.ResourceRepresentation, error) {
	resourceRepresentation, err := r.client.CreateResourceClient(ctx, token, realm, resource)
	if err != nil {
		return nil, err
	}
	return resourceRepresentation, nil
}

func (r *IdentityProvider) UpdateResourceClient(ctx context.Context, token, realm string, resource gocloak.ResourceRepresentation) error {
	if err := r.client.UpdateResourceClient(ctx, token, realm, resource); err != nil {
		return err
	}
	return nil
}

func (r *IdentityProvider) DeleteResourceClient(ctx context.Context, token, realm, resourceID string) error {
	if err := r.client.DeleteResourceClient(ctx, token, realm, resourceID); err != nil {
		return err
	}
	return nil
}

func (r *IdentityProvider) GetScope(ctx context.Context, token, realm, idOfClient, scopeID string) (*gocloak.ScopeRepresentation, error) {
	scopeRepresentation, err := r.client.GetScope(ctx, token, realm, idOfClient, scopeID)
	if err != nil {
		return nil, err
	}
	return scopeRepresentation, nil
}

func (r *IdentityProvider) GetScopes(ctx context.Context, token, realm, idOfClient string, params gocloak.GetScopeParams) ([]*gocloak.ScopeRepresentation, error) {
	scopeRepresentation, err := r.client.GetScopes(ctx, token, realm, idOfClient, params)
	if err != nil {
		return nil, err
	}
	return scopeRepresentation, nil
}

func (r *IdentityProvider) CreateScope(ctx context.Context, token, realm, idOfClient string, scope gocloak.ScopeRepresentation) (*gocloak.ScopeRepresentation, error) {
	scopeRepresentation, err := r.client.CreateScope(ctx, token, realm, idOfClient, scope)
	if err != nil {
		return nil, err
	}
	return scopeRepresentation, nil
}

func (r *IdentityProvider) UpdateScope(ctx context.Context, token, realm, idOfClient string, resource gocloak.ScopeRepresentation) error {
	if err := r.client.UpdateScope(ctx, token, realm, idOfClient, resource); err != nil {
		return err
	}
	return nil
}
func (r *IdentityProvider) DeleteScope(ctx context.Context, token, realm, idOfClient, scopeID string) error {
	if err := r.client.DeleteScope(ctx, token, realm, idOfClient, scopeID); err != nil {
		return err
	}
	return nil
}

func (r *IdentityProvider) GetPolicy(ctx context.Context, token, realm, idOfClient, policyID string) (*gocloak.PolicyRepresentation, error) {
	policyRepresentation, err := r.client.GetPolicy(ctx, token, realm, idOfClient, policyID)
	if err != nil {
		return nil, err
	}
	return policyRepresentation, nil
}

func (r *IdentityProvider) GetPolicies(ctx context.Context, token, realm, idOfClient string, params gocloak.GetPolicyParams) ([]*gocloak.PolicyRepresentation, error) {
	policyRepresentation, err := r.client.GetPolicies(ctx, token, realm, idOfClient, params)
	if err != nil {
		return nil, err
	}
	return policyRepresentation, nil
}

func (r *IdentityProvider) CreatePolicy(ctx context.Context, token, realm, idOfClient string, policy gocloak.PolicyRepresentation) (*gocloak.PolicyRepresentation, error) {
	policyRepresentation, err := r.client.CreatePolicy(ctx, token, realm, idOfClient, policy)
	if err != nil {
		return nil, err
	}
	return policyRepresentation, nil
}

func (r *IdentityProvider) UpdatePolicy(ctx context.Context, token, realm, idOfClient string, policy gocloak.PolicyRepresentation) error {
	if err := r.client.UpdatePolicy(ctx, token, realm, idOfClient, policy); err != nil {
		return err
	}
	return nil
}

func (r *IdentityProvider) DeletePolicy(ctx context.Context, token, realm, idOfClient, policyID string) error {
	if err := r.client.DeletePolicy(ctx, token, realm, idOfClient, policyID); err != nil {
		return err
	}
	return nil
}

func (r *IdentityProvider) GetResourcePolicy(ctx context.Context, token, realm, permissionID string) (*gocloak.ResourcePolicyRepresentation, error) {
	resourcePolicyRepresentation, err := r.client.GetResourcePolicy(ctx, token, realm, permissionID)
	if err != nil {
		return nil, err
	}
	return resourcePolicyRepresentation, nil
}

func (r *IdentityProvider) GetResourcePolicies(ctx context.Context, token, realm string, params gocloak.GetResourcePoliciesParams) ([]*gocloak.ResourcePolicyRepresentation, error) {
	resourcePolicyRepresentation, err := r.client.GetResourcePolicies(ctx, token, realm, params)
	if err != nil {
		return nil, err
	}
	return resourcePolicyRepresentation, nil
}

func (r *IdentityProvider) CreateResourcePolicy(ctx context.Context, token, realm, resourceID string, policy gocloak.ResourcePolicyRepresentation) (*gocloak.ResourcePolicyRepresentation, error) {
	resourcePolicyRepresentation, err := r.client.CreateResourcePolicy(ctx, token, realm, resourceID, policy)
	if err != nil {
		return nil, err
	}
	return resourcePolicyRepresentation, nil
}

func (r *IdentityProvider) UpdateResourcePolicy(ctx context.Context, token, realm, permissionID string, policy gocloak.ResourcePolicyRepresentation) error {
	if err := r.client.UpdateResourcePolicy(ctx, token, realm, permissionID, policy); err != nil {
		return err
	}
	return nil
}

func (r *IdentityProvider) DeleteResourcePolicy(ctx context.Context, token, realm, permissionID string) error {
	if err := r.client.DeleteResourcePolicy(ctx, token, realm, permissionID); err != nil {
		return err
	}
	return nil
}

func (r *IdentityProvider) GetPermission(ctx context.Context, token, realm, idOfClient, permissionID string) (*gocloak.PermissionRepresentation, error) {
	permissionRepresentation, err := r.client.GetPermission(ctx, token, realm, idOfClient, permissionID)
	if err != nil {
		return nil, err
	}
	return permissionRepresentation, nil
}

func (r *IdentityProvider) GetPermissions(ctx context.Context, token, realm, idOfClient string, params gocloak.GetPermissionParams) ([]*gocloak.PermissionRepresentation, error) {
	permissionRepresentation, err := r.client.GetPermissions(ctx, token, realm, idOfClient, params)
	if err != nil {
		return nil, err
	}
	return permissionRepresentation, nil
}

func (r *IdentityProvider) GetPermissionResources(ctx context.Context, token, realm, idOfClient, permissionID string) ([]*gocloak.PermissionResource, error) {
	permissionResource, err := r.client.GetPermissionResources(ctx, token, realm, idOfClient, permissionID)
	if err != nil {
		return nil, err
	}
	return permissionResource, nil
}

func (r *IdentityProvider) GetPermissionScopes(ctx context.Context, token, realm, idOfClient, permissionID string) ([]*gocloak.PermissionScope, error) {
	permissionScope, err := r.client.GetPermissionScopes(ctx, token, realm, idOfClient, permissionID)
	if err != nil {
		return nil, err
	}
	return permissionScope, nil
}

func (r *IdentityProvider) GetDependentPermissions(ctx context.Context, token, realm, idOfClient, policyID string) ([]*gocloak.PermissionRepresentation, error) {
	permissionRepresentation, err := r.client.GetDependentPermissions(ctx, token, realm, idOfClient, policyID)
	if err != nil {
		return nil, err
	}
	return permissionRepresentation, nil
}

func (r *IdentityProvider) CreatePermission(ctx context.Context, token, realm, idOfClient string, permission gocloak.PermissionRepresentation) (*gocloak.PermissionRepresentation, error) {
	permissionRepresentation, err := r.client.CreatePermission(ctx, token, realm, idOfClient, permission)
	if err != nil {
		return nil, err
	}
	return permissionRepresentation, nil
}

func (r *IdentityProvider) UpdatePermission(ctx context.Context, token, realm, idOfClient string, permission gocloak.PermissionRepresentation) error {
	if err := r.client.UpdatePermission(ctx, token, realm, idOfClient, permission); err != nil {
		return err
	}
	return nil
}

func (r *IdentityProvider) DeletePermission(ctx context.Context, token, realm, idOfClient, permissionID string) error {
	if err := r.client.DeletePermission(ctx, token, realm, idOfClient, permissionID); err != nil {
		return err
	}
	return nil
}

func (r *IdentityProvider) CreatePermissionTicket(ctx context.Context, token, realm string, permissions []gocloak.CreatePermissionTicketParams) (*gocloak.PermissionTicketResponseRepresentation, error) {
	permissionTicketResponseRepresentation, err := r.client.CreatePermissionTicket(ctx, token, realm, permissions)
	if err != nil {
		return nil, err
	}
	return permissionTicketResponseRepresentation, nil
}

func (r *IdentityProvider) GrantUserPermission(ctx context.Context, token, realm string, permission gocloak.PermissionGrantParams) (*gocloak.PermissionGrantResponseRepresentation, error) {
	permissionGrantResponseRepresentation, err := r.client.GrantUserPermission(ctx, token, realm, permission)
	if err != nil {
		return nil, err
	}
	return permissionGrantResponseRepresentation, nil
}

func (r *IdentityProvider) UpdateUserPermission(ctx context.Context, token, realm string, permission gocloak.PermissionGrantParams) (*gocloak.PermissionGrantResponseRepresentation, error) {
	permissionGrantResponseRepresentation, err := r.client.UpdateUserPermission(ctx, token, realm, permission)
	if err != nil {
		return nil, err
	}
	return permissionGrantResponseRepresentation, nil
}

func (r *IdentityProvider) GetUserPermissions(ctx context.Context, token, realm string, params gocloak.GetUserPermissionParams) ([]*gocloak.PermissionGrantResponseRepresentation, error) {
	permissionGrantResponseRepresentation, err := r.client.GetUserPermissions(ctx, token, realm, params)
	if err != nil {
		return nil, err
	}
	return permissionGrantResponseRepresentation, nil
}

func (r *IdentityProvider) DeleteUserPermission(ctx context.Context, token, realm, ticketID string) error {
	if err := r.client.DeleteUserPermission(ctx, token, realm, ticketID); err != nil {
		return err
	}
	return nil
}

func (r *IdentityProvider) GetCredentialRegistrators(ctx context.Context, token, realm string) ([]string, error) {
	msgs, err := r.client.GetCredentialRegistrators(ctx, token, realm)
	if err != nil {
		return nil, err
	}
	return msgs, nil
}

func (r *IdentityProvider) GetConfiguredUserStorageCredentialTypes(ctx context.Context, token, realm, userID string) ([]string, error) {
	msgs, err := r.client.GetConfiguredUserStorageCredentialTypes(ctx, token, realm, userID)
	if err != nil {
		return nil, err
	}
	return msgs, nil
}

func (r *IdentityProvider) GetCredentials(ctx context.Context, token, realm, UserID string) ([]*gocloak.CredentialRepresentation, error) {
	credentialRepresentation, err := r.client.GetCredentials(ctx, token, realm, UserID)
	if err != nil {
		return nil, err
	}
	return credentialRepresentation, nil
}

func (r *IdentityProvider) DeleteCredentials(ctx context.Context, token, realm, UserID, CredentialID string) error {
	if err := r.client.DeleteCredentials(ctx, token, realm, UserID, CredentialID); err != nil {
		return err
	}
	return nil
}

func (r *IdentityProvider) UpdateCredentialUserLabel(ctx context.Context, token, realm, userID, credentialID, userLabel string) error {
	if err := r.client.UpdateCredentialUserLabel(ctx, token, realm, userID, credentialID, userLabel); err != nil {
		return err
	}
	return nil
}

func (r *IdentityProvider) DisableAllCredentialsByType(ctx context.Context, token, realm, userID string, types []string) error {
	if err := r.client.DisableAllCredentialsByType(ctx, token, realm, userID, types); err != nil {
		return err
	}
	return nil
}

func (r *IdentityProvider) MoveCredentialBehind(ctx context.Context, token, realm, userID, credentialID, newPreviousCredentialID string) error {
	if err := r.client.MoveCredentialBehind(ctx, token, realm, userID, credentialID, newPreviousCredentialID); err != nil {
		return err
	}
	return nil
}

func (r *IdentityProvider) MoveCredentialToFirst(ctx context.Context, token, realm, userID, credentialID string) error {
	if err := r.client.MoveCredentialToFirst(ctx, token, realm, userID, credentialID); err != nil {
		return err
	}
	return nil
}

func (r *IdentityProvider) GetAuthenticationFlows(ctx context.Context, token, realm string) ([]*gocloak.AuthenticationFlowRepresentation, error) {
	authenticationFlowRepresentation, err := r.client.GetAuthenticationFlows(ctx, token, realm)
	if err != nil {
		return nil, err
	}
	return authenticationFlowRepresentation, nil
}

func (r *IdentityProvider) GetAuthenticationFlow(ctx context.Context, token, realm string, authenticationFlowID string) (*gocloak.AuthenticationFlowRepresentation, error) {
	authenticationFlowRepresentation, err := r.client.GetAuthenticationFlow(ctx, token, realm, authenticationFlowID)
	if err != nil {
		return nil, err
	}
	return authenticationFlowRepresentation, nil
}

func (r *IdentityProvider) CreateAuthenticationFlow(ctx context.Context, token, realm string, flow gocloak.AuthenticationFlowRepresentation) error {
	if err := r.client.CreateAuthenticationFlow(ctx, token, realm, flow); err != nil {
		return err
	}
	return nil
}

func (r *IdentityProvider) UpdateAuthenticationFlow(ctx context.Context, token, realm string, flow gocloak.AuthenticationFlowRepresentation, authenticationFlowID string) (*gocloak.AuthenticationFlowRepresentation, error) {
	authenticationFlowRepresentation, err := r.client.UpdateAuthenticationFlow(ctx, token, realm, flow, authenticationFlowID)
	if err != nil {
		return nil, err
	}
	return authenticationFlowRepresentation, nil
}

func (r *IdentityProvider) DeleteAuthenticationFlow(ctx context.Context, token, realm, flowID string) error {
	if err := r.client.DeleteAuthenticationFlow(ctx, token, realm, flowID); err != nil {
		return err
	}
	return nil
}

func (r *IdentityProvider) CreateIdentityProvider(ctx context.Context, token, realm string, providerRep gocloak.IdentityProviderRepresentation) (string, error) {
	msg, err := r.client.CreateIdentityProvider(ctx, token, realm, providerRep)
	if err != nil {
		return "", err
	}
	return msg, nil
}

func (r *IdentityProvider) GetIdentityProvider(ctx context.Context, token, realm, alias string) (*gocloak.IdentityProviderRepresentation, error) {
	identityProviderRepresentation, err := r.client.GetIdentityProvider(ctx, token, realm, alias)
	if err != nil {
		return nil, err
	}
	return identityProviderRepresentation, nil
}

func (r *IdentityProvider) GetIdentityProviders(ctx context.Context, token, realm string) ([]*gocloak.IdentityProviderRepresentation, error) {
	identityProviderRepresentation, err := r.client.GetIdentityProviders(ctx, token, realm)
	if err != nil {
		return nil, err
	}
	return identityProviderRepresentation, nil
}

func (r *IdentityProvider) UpdateIdentityProvider(ctx context.Context, token, realm, alias string, providerRep gocloak.IdentityProviderRepresentation) error {
	if err := r.client.UpdateIdentityProvider(ctx, token, realm, alias, providerRep); err != nil {
		return err
	}
	return nil
}

func (r *IdentityProvider) DeleteIdentityProvider(ctx context.Context, token, realm, alias string) error {
	if err := r.client.DeleteIdentityProvider(ctx, token, realm, alias); err != nil {
		return err
	}
	return nil
}

func (r *IdentityProvider) CreateIdentityProviderMapper(ctx context.Context, token, realm, alias string, mapper gocloak.IdentityProviderMapper) (string, error) {
	msg, err := r.client.CreateIdentityProviderMapper(ctx, token, realm, alias, mapper)
	if err != nil {
		return "", nil
	}
	return msg, nil
}

func (r *IdentityProvider) GetIdentityProviderMapper(ctx context.Context, token string, realm string, alias string, mapperID string) (*gocloak.IdentityProviderMapper, error) {
	identityProviderMapper, err := r.client.GetIdentityProviderMapper(ctx, token, realm, alias, mapperID)
	if err != nil {
		return nil, err
	}
	return identityProviderMapper, nil
}

func (r *IdentityProvider) CreateUserFederatedIdentity(ctx context.Context, token, realm, userID, providerID string, federatedIdentityRep gocloak.FederatedIdentityRepresentation) error {
	if err := r.client.CreateUserFederatedIdentity(ctx, token, realm, userID, providerID, federatedIdentityRep); err != nil {
		return err
	}
	return nil
}

func (r *IdentityProvider) GetUserFederatedIdentities(ctx context.Context, token, realm, userID string) ([]*gocloak.FederatedIdentityRepresentation, error) {
	federatedIdentityRepresentation, err := r.client.GetUserFederatedIdentities(ctx, token, realm, userID)
	if err != nil {
		return nil, err
	}
	return federatedIdentityRepresentation, nil
}

func (r *IdentityProvider) DeleteUserFederatedIdentity(ctx context.Context, token, realm, userID, providerID string) error {
	if err := r.client.DeleteUserFederatedIdentity(ctx, token, realm, userID, providerID); err != nil {
		return err
	}
	return nil
}

func (r *IdentityProvider) GetEvents(ctx context.Context, token string, realm string, params gocloak.GetEventsParams) ([]*gocloak.EventRepresentation, error) {
	eventRepresentation, err := r.client.GetEvents(ctx, token, realm, params)
	if err != nil {
		return nil, err
	}
	return eventRepresentation, nil
}
