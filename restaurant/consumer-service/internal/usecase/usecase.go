package usecase

import "github.com/JIeeiroSst/consumer-service/internal/repository"

type Usecase struct {
	Consumers
}

type Dependency struct {
	Repos *repository.Repository
}

func NewUsecase(deps Dependency) *Usecase {
	return &Usecase{
		Consumers: NewConsumerUsecase(deps.Repos.Consumers),
	}
}
