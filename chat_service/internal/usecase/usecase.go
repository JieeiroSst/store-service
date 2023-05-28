package usecase

import (
	"github.com/JIeeiroSst/chat-service/internal/repository"
	"github.com/JIeeiroSst/chat-service/pkg/cache"
	"github.com/JIeeiroSst/chat-service/pkg/snowflake"
)

type Usecase struct {
	Messages
}

type Dependency struct {
	CacheHelper cache.CacheHelper
	Repo        *repository.Repositories
	Snowflake   snowflake.SnowflakeData
}

func NewUsecase(deps Dependency) *Usecase {
	messageRepo :=  NewMessageUsecase(deps.Repo, deps.CacheHelper, deps.Snowflake) 
	return &Usecase{
		Messages: messageRepo,
	}
}
