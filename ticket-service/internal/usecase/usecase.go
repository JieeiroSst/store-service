package usecase

import (
	"github.com/JIeeiroSst/ticket-service/internal/repository"
	"github.com/JIeeiroSst/utils/cache/expire"
)

type Usecase struct {
	Tickets
	Invoices
}

type Dependency struct {
	Repos           *repository.Repository
	CacheHelper     expire.CacheHelper
	UnidocSerectKey string
}

func NewUsecase(deps Dependency) *Usecase {
	tickets := NewTicketsUsecase(deps.Repos)
	invoices := NewInvoicesUsecase(deps.Repos, deps.UnidocSerectKey)

	return &Usecase{
		Tickets:  tickets,
		Invoices: invoices,
	}
}
