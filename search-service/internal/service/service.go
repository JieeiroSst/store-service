package service

import "github.com/JIeeiroSst/search-service/internal/repository"

type Service struct {
	Tasks
}

type Dependency struct {
	Repos *repository.Repositories
}

func NewUsecase(deps Dependency) *Service {
	taskService := NewTask(deps.Repos.TaskRepository)

	return &Service{
		Tasks: taskService,
	}
}
