package repository

import "github.com/JIeeiroSst/partner-service/internal/core/domain"

func (m *DB) CreateProject(userID string, Project domain.Project) error {
	return nil
}

func (m *DB) ReadProject(id string) (*domain.Project, error) {
	return nil, nil
}

func (m *DB) ReadProjects(pagination domain.Pagination) (*domain.Pagination, error) {
	var project []*domain.Project

	m.db.Scopes(paginate(project, "", &pagination, m.db)).Find(&project)
	pagination.Rows = project

	return &pagination, nil
}

func (m *DB) UpdateProject(id string, Project domain.Project) error {
	return nil
}

func (m *DB) DeleteProject(id string) error {
	return nil
}
