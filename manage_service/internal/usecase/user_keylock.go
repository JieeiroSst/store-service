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
	GetClients(ctx context.Context, user dto.Client) ([]*keycloak.Client, error)
	Login(ctx context.Context, clientID, clientSecret, realm, username, password string) (*keycloak.JWT, error)
	LoginOtp(ctx context.Context, clientID, clientSecret, realm, username, password, totp string) (*keycloak.JWT, error)
	Logout(ctx context.Context, clientID, clientSecret, realm, refreshToken string) error
	LoginClient(ctx context.Context, clientID, clientSecret, realm string) (*keycloak.JWT, error)
	RefreshToken(ctx context.Context, refreshToken, clientID, clientSecret, realm string) (*keycloak.JWT, error)
	GetUserInfo(ctx context.Context, accessToken, realm string) (*keycloak.UserInfo, error)
	SetPassword(ctx context.Context, token, userID, realm, password string, temporary bool) error
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

func (u *UserKeycloakUsecase) IntrospectToken(ctx context.Context, token dto.IntrospectToken) (*[]keycloak.ResourcePermission, error) {
	tokenModel := model.IntrospectToken{
		Token:        token.Token,
		Realm:        token.Realm,
		ClientID:     token.ClientID,
		ClientSecret: token.ClientSecret,
	}
	resourcePermission, err := u.UserKeycloakRepo.IntrospectToken(ctx, tokenModel)
	if err != nil {
		return nil, err
	}
	return resourcePermission, nil
}

func (u *UserKeycloakUsecase) GetClients(ctx context.Context, user dto.Client) ([]*keycloak.Client, error) {
	userModel := model.Client{
		Token:      user.Token,
		ClientName: user.ClientName,
		Realm:      user.Realm,
	}
	client, err := u.UserKeycloakRepo.GetClients(ctx, userModel)
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
