package usecase

import (
	"context"

	"github.com/JIeeiroSst/manage-service/internal/dto"
	"github.com/JIeeiroSst/manage-service/internal/model"
	"github.com/JIeeiroSst/manage-service/internal/repository"
)

type UserKeyclock interface {
	LoginAdmin(ctx context.Context, user dto.Login) (*dto.Token, error)
	GetTokenUser(ctx context.Context, realm string) (*dto.TokenInfo, error)
	CreateUser(ctx context.Context, user dto.CreateUser) error
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
