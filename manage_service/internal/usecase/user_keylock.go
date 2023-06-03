package usecase

import (
	"context"

	"github.com/JIeeiroSst/manage-service/internal/dto"
	"github.com/JIeeiroSst/manage-service/internal/model"
	"github.com/JIeeiroSst/manage-service/internal/repository"
	keycloak "github.com/Nerzal/gocloak/v13"
)

type UserKeyclock interface {
	LoginAdmin(ctx context.Context, user dto.LoginAdmin) (*dto.Token, error)
	GetTokenUser(ctx context.Context, realm string) (*dto.TokenInfo, error)
	CreateUser(ctx context.Context, user dto.CreateUser) error
	IntrospectToken(ctx context.Context, token dto.IntrospectToken) (*[]keycloak.ResourcePermission, error)
	GetClients(ctx context.Context, user dto.Client) ([]*keycloak.Client, error)
	Login(ctx context.Context, user dto.Login) (*keycloak.JWT, error)
	LoginOtp(ctx context.Context, user dto.LoginOTP) (*keycloak.JWT, error)
	Logout(ctx context.Context, user dto.Logout) error
	LoginClient(ctx context.Context, user dto.LoginClient) (*keycloak.JWT, error)
	RefreshToken(ctx context.Context, user dto.RefreshToken) (*keycloak.JWT, error)
	GetUserInfo(ctx context.Context, user dto.UserInfo) (*keycloak.UserInfo, error)
	SetPassword(ctx context.Context, user dto.SetPassword) error
}

type UserKeycloakUsecase struct {
	UserKeycloakRepo repository.UserKeycloak
}

func NewUserKeycloakUsecase(UserKeycloakRepo repository.UserKeycloak) *UserKeycloakUsecase {
	return &UserKeycloakUsecase{
		UserKeycloakRepo: UserKeycloakRepo,
	}
}

func (u *UserKeycloakUsecase) LoginAdmin(ctx context.Context, user dto.LoginAdmin) (*dto.Token, error) {
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

func (u *UserKeycloakUsecase) Login(ctx context.Context, user dto.Login) (*keycloak.JWT, error) {
	token, err := u.UserKeycloakRepo.Login(ctx, user.ClientID, user.ClientSecret, user.Realm, user.Username, user.Password)
	if err != nil {
		return nil, err
	}
	return token, nil
}
func (u *UserKeycloakUsecase) LoginOtp(ctx context.Context, user dto.LoginOTP) (*keycloak.JWT, error) {
	token, err := u.UserKeycloakRepo.LoginOtp(ctx, user.ClientID, user.ClientSecret, user.Realm, user.Username, user.Password, user.OTP)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (u *UserKeycloakUsecase) Logout(ctx context.Context, user dto.Logout) error {
	if err := u.UserKeycloakRepo.Logout(ctx, user.ClientID, user.ClientSecret, user.Realm, user.RefreshToken); err != nil {
		return err
	}
	return nil
}

func (u *UserKeycloakUsecase) LoginClient(ctx context.Context, user dto.LoginClient) (*keycloak.JWT, error) {
	token, err := u.UserKeycloakRepo.LoginClient(ctx, user.ClientID, user.ClientSecret, user.Realm)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (u *UserKeycloakUsecase) RefreshToken(ctx context.Context, user dto.RefreshToken) (*keycloak.JWT, error) {
	token, err := u.UserKeycloakRepo.RefreshToken(ctx, user.RefreshToken, user.ClientID, user.ClientSecret, user.Realm)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (u *UserKeycloakUsecase) GetUserInfo(ctx context.Context, user dto.UserInfo) (*keycloak.UserInfo, error) {
	userInfo, err := u.UserKeycloakRepo.GetUserInfo(ctx, user.AccessToken, user.Realm)
	if err != nil {
		return nil, err
	}
	return userInfo, nil
}

func (u *UserKeycloakUsecase) SetPassword(ctx context.Context, user dto.SetPassword) error {
	if err := u.UserKeycloakRepo.SetPassword(ctx, user.Token, user.UserID, user.Realm, user.Password, user.Temporary); err != nil {
		return err
	}
	return nil
}
