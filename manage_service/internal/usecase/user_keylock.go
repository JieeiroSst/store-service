package usecase

import "github.com/JIeeiroSst/manage-service/internal/repository"

type UserKeyclock interface {
}

type UserKeycloakUsecase struct {
	UserKeycloakRepo repository.UserKeycloak
}

func NewUserKeycloakUsecase(UserKeycloakRepo repository.UserKeycloak) *UserKeycloakUsecase {
	return &UserKeycloakUsecase{
		UserKeycloakRepo: UserKeycloakRepo,
	}
}
