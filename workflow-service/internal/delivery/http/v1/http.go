package v1

import (
	"github.com/JIeeiroSst/workflow-service/internal/usecase"
	"github.com/go-chi/chi/v5"
)

type Http struct {
	Usecase *usecase.Usecase
}

func NewHttpV1(Usecase *usecase.Usecase) *Http {
	return &Http{
		Usecase: Usecase,
	}
}

func (u *Http) SetupRoutes(router chi.Router) {

}
