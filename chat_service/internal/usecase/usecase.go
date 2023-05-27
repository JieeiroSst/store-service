package usecase

import (
	"github.com/JIeeiroSst/chat-service/internal/repository"
	"github.com/JIeeiroSst/chat-service/pkg/cache"
)

type Usecase struct {
	Messages
}

type Dependency struct {
	CacheHelper cache.CacheHelper
	MessageRepo repository.MessageRepo
}

func NewUsecase(deps Dependency) *Usecase {
	return &Usecase{
		Messages: NewMessageUsecase(deps.MessageRepo, deps.CacheHelper),
	}
}
