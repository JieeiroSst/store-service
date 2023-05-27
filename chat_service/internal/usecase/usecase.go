package usecase

import "github.com/JIeeiroSst/chat-service/pkg/cache"

type Usecase struct {
}

type Dependency struct {
	CacheHelper cache.CacheHelper
}

func NewUsecase(deps Dependency) *Usecase {
	return &Usecase{}
}
