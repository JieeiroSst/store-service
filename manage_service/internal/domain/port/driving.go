package port

import (
	"context"

	"github.com/JIeeiroSst/manage-service/internal/domain/model"
	"github.com/Nerzal/gocloak/v13"
)

type AuthUsecase interface {
	LoginAdmin(ctx context.Context, user model.LoginAdmin) (*model.Token, error)
	GetTokenUser(ctx context.Context, realm string) (*model.TokenInfo, error)
	CreateUser(ctx context.Context, user model.CreateUser) error
	IntrospectToken(ctx context.Context, token model.IntrospectToken) (*[]gocloak.ResourcePermission, error)
	GetClients(ctx context.Context, client model.Client) ([]*gocloak.Client, error)
	Login(ctx context.Context, user model.Login) (*gocloak.JWT, error)
	LoginOtp(ctx context.Context, user model.LoginOTP) (*gocloak.JWT, error)
	Logout(ctx context.Context, user model.Logout) error
	LoginClient(ctx context.Context, user model.LoginClient) (*gocloak.JWT, error)
	RefreshToken(ctx context.Context, user model.RefreshToken) (*gocloak.JWT, error)
	GetUserInfo(ctx context.Context, user model.UserInfo) (*gocloak.UserInfo, error)
	SetPassword(ctx context.Context, user model.SetPassword) error
}

type KeycloakUsecase interface {
	IdentityProvider
	AdminToken(ctx context.Context) (string, error)
}
