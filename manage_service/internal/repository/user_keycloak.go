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

