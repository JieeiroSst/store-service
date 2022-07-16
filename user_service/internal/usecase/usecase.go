package usecase

import (
	"github.com/JIeeiroSst/user-service/internal/repository"
	"github.com/JIeeiroSst/user-service/pkg/hash"
	"github.com/JIeeiroSst/user-service/pkg/snowflake"
	"github.com/JIeeiroSst/user-service/pkg/token"
)

type Usecase struct {
	Users
}

type Dependency struct {
	Repos     *repository.Repository
	Snowflake snowflake.SnowflakeData
	Hash      hash.Hash
	Token     token.Tokens
}

func NewUsecase(deps Dependency) *Usecase {
	return &Usecase{
		Users: NewUsercase(deps.Repos.Users, deps.Snowflake, deps.Hash, deps.Token),
	}
}
