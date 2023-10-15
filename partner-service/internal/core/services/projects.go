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

func (u *ProjectService) CreateProject(userID string, Project domain.Project) error

func (u *ProjectService) ReadProject(id string) (*domain.Project, error)

func (u *ProjectService) ReadProjects(pagination domain.Pagination) (*domain.Pagination, error)

func (u *ProjectService) UpdateProject(id string, Project domain.Project) error

func (u *ProjectService) DeleteProject(id string) error
