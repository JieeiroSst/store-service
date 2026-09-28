package application

import (
	"context"

	"github.com/JIeeiroSst/manage-service/internal/domain/model"
	"github.com/JIeeiroSst/manage-service/internal/domain/port"
	"github.com/Nerzal/gocloak/v13"
)

type AuthService struct {
	idp port.IdentityProvider
}

func NewAuthService(idp port.IdentityProvider) port.AuthUsecase {
	return &AuthService{idp: idp}
}

func (s *AuthService) LoginAdmin(ctx context.Context, user model.LoginAdmin) (*model.Token, error) {
	return s.idp.LoginAdmin(ctx, user)
}

func (s *AuthService) GetTokenUser(ctx context.Context, realm string) (*model.TokenInfo, error) {
	return s.idp.GetTokenUser(ctx, realm)
}

func (s *AuthService) CreateUser(ctx context.Context, user model.CreateUser) error {
	return s.idp.CreateUser(ctx, user)
}

func (s *AuthService) IntrospectToken(ctx context.Context, token model.IntrospectToken) (*[]gocloak.ResourcePermission, error) {
	return s.idp.IntrospectToken(ctx, token)
}

func (s *AuthService) GetClients(ctx context.Context, client model.Client) ([]*gocloak.Client, error) {
	return s.idp.GetClients(ctx, client)
}

func (s *AuthService) Login(ctx context.Context, user model.Login) (*gocloak.JWT, error) {
	return s.idp.Login(ctx, user.ClientID, user.ClientSecret, user.Realm, user.Username, user.Password)
}

func (s *AuthService) LoginOtp(ctx context.Context, user model.LoginOTP) (*gocloak.JWT, error) {
	return s.idp.LoginOtp(ctx, user.ClientID, user.ClientSecret, user.Realm, user.Username, user.Password, user.OTP)
}

func (s *AuthService) Logout(ctx context.Context, user model.Logout) error {
	return s.idp.Logout(ctx, user.ClientID, user.ClientSecret, user.Realm, user.RefreshToken)
}

func (s *AuthService) LoginClient(ctx context.Context, user model.LoginClient) (*gocloak.JWT, error) {
	return s.idp.LoginClient(ctx, user.ClientID, user.ClientSecret, user.Realm)
}

func (s *AuthService) RefreshToken(ctx context.Context, user model.RefreshToken) (*gocloak.JWT, error) {
	return s.idp.RefreshToken(ctx, user.RefreshToken, user.ClientID, user.ClientSecret, user.Realm)
}

func (s *AuthService) GetUserInfo(ctx context.Context, user model.UserInfo) (*gocloak.UserInfo, error) {
	return s.idp.GetUserInfo(ctx, user.AccessToken, user.Realm)
}

func (s *AuthService) SetPassword(ctx context.Context, user model.SetPassword) error {
	return s.idp.SetPassword(ctx, user.Token, user.UserID, user.Realm, user.Password, user.Temporary)
}
