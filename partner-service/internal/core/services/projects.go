package services

import (
	"github.com/JIeeiroSst/partner-service/internal/core/domain"
	"github.com/JIeeiroSst/partner-service/internal/core/ports"
)

type ProjectService struct {
	repo ports.ProjectRepository
}

func NewProjectService(repo ports.ProjectRepository) *ProjectService {
	return &ProjectService{
		repo: repo,
	}
}

func (u *ProjectService) CreateProject(userID string, Project domain.Project) error {
	return u.repo.CreateProject(userID, Project)
}

func (u *ProjectService) ReadProject(id string) (*domain.Project, error) {
	return u.repo.ReadProject(id)
}

func (u *ProjectService) ReadProjects(pagination domain.Pagination) (*domain.Pagination, error) {
	return u.repo.ReadProjects(pagination)
}

func (u *ProjectService) UpdateProject(id string, Project domain.Project) error {
	return u.repo.UpdateProject(id, Project)
}

func (u *ProjectService) DeleteProject(id string) error {
	return u.repo.DeleteProject(id)
}
