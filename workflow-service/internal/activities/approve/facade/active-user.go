package facade

import "github.com/JIeeiroSst/workflow-service/internal/repository"

type ActiveUser interface {
	InsertActiveUser(user ActiveUser) error
}

type ActiveUserFace struct {
	repository *repository.Repositories
}

func NewActiveUserFace(repository *repository.Repositories) *ActiveUserFace {
	return &ActiveUserFace{
		repository: repository,
	}
}

func (u *ActiveUserFace) InsertActiveUser(user ActiveUser) error {
	return nil
}
