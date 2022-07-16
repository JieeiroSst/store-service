package usecase

import (
	"github.com/JIeeiroSst/post-service/internal/repository"
	"github.com/JIeeiroSst/post-service/pkg/snowflake"
)

type Usecase struct {
}

type Dependency struct {
	Repos     *repository.Repository
	Snowflake snowflake.SnowflakeData
}

func NewUsecase(deps Dependency) *Usecase {
	return &Usecase{}
}
