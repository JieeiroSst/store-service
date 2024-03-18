package usecase

import (
	"github.com/JIeerioSst/subsidies-service/internal/repository"
	"github.com/go-redis/redis/v8"
)

type Usecase struct {
}

type Dependency struct {
	Repos       *repository.Repositories
	CacheHelper *redis.Client
}

func NewUsecase(deps Dependency) *Usecase {
	return &Usecase{}
}
