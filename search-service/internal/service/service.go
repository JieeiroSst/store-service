package service

import "github.com/JIeeiroSst/search-service/internal/repository"

type Usecase struct {
	Tasks
}

type Dependency struct {
	Repos *repository.Repositories
}

func NewUsecase(deps Dependency) *Usecase {
	taskService := NewTask(deps.Repos.TaskRepository)

	return &Usecase{
		Tasks: taskService,
	}
}
