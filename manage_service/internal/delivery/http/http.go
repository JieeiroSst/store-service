package http

import (
	"github.com/JIeeiroSst/manage-service/config"
	"github.com/JIeeiroSst/manage-service/internal/usecase"
)

type Http struct {
	Usecase *usecase.Usecase
	Config  *config.Config
}

func NewHttp(Usecase *usecase.Usecase, Config *config.Config) *Http {
	return &Http{
		Usecase: Usecase,
		Config:  Config,
	}
}
