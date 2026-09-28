package application

import (
	"context"

	"github.com/JIeeiroSst/basket-service/internal/domain/model"
	"github.com/JIeeiroSst/basket-service/internal/domain/port"
)

type userService struct {
	users port.UserDirectory
}

func NewUserService(users port.UserDirectory) port.UserUsecase {
	return &userService{users: users}
}

func (s *userService) GetUser(ctx context.Context, id int) (*model.User, error) {
	return s.users.GetByID(ctx, id)
}
