package application

import (
	"context"

	"github.com/JIeeiroSst/basket-service/internal/domain/model"
	"github.com/JIeeiroSst/basket-service/internal/domain/port"
)

type userService struct {
	repo port.UserRepository
}

func NewUserService(repo port.UserRepository) port.UserUsecase {
	return &userService{repo: repo}
}

func (s *userService) GetUser(ctx context.Context, id int) (*model.User, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *userService) ListUsers(ctx context.Context) ([]model.User, error) {
	return s.repo.List(ctx)
}
