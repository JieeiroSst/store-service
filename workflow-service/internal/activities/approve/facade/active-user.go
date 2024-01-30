package facade

import (
	"github.com/JIeeiroSst/workflow-service/dto"
	"github.com/JIeeiroSst/workflow-service/internal/repository"
)

type ActiveUser interface {
	InsertActiveUser(user dto.ActiveUser, id string) error
}

type ActiveUserFace struct {
	repository *repository.Repositories
}

func NewActiveUserFace(repository *repository.Repositories) *ActiveUserFace {
	return &ActiveUserFace{
		repository: repository,
	}
}

func (u *ActiveUserFace) InsertActiveUser(user dto.ActiveUser, id string) error {
	userModel := dto.FormatActiveUser(user)

	if err := u.repository.ActiveUsers.InsertActiveUser(userModel, id); err != nil {
		return err
	}

	return nil
}
