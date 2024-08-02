package usecase

import (
	"github.com/JIeeiroSst/offer-range-service/internal/repository"
)

type Usecase struct {
}

type Dependency struct {
	Repos *repository.Repository
}
