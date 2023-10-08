package services

import "github.com/JIeeiroSst/partner-service/internal/core/ports"

type ProjectService struct {
	repo ports.ProjectRepository
}

func NewProjectService(repo ports.ProjectRepository) *ProjectService {
	return &ProjectService{
		repo: repo,
	}
}
