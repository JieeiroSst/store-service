package usecase

import "github.com/JIeeiroSst/delivery-service/internal/repository"

type Usecase struct {
	Deliveries
}

type Dependency struct {
	Repos *repository.Repository
}

func NewUsecase(deps Dependency) *Usecase {
	return &Usecase{
		Deliveries: NewDeliveryRepository(deps.Repos.Deliveries),
	}
}
