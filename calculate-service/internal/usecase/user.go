package usecase

import "github.com/JIeeiroSst/calculate-service/internal/repository"

type Users interface {
}

type UserUsecase struct {
	Repo *repository.Repository
}

func NewUserUsecase(repo *repository.Repository) *UserUsecase {
	return &UserUsecase{
		Repo: repo,
	}
}
