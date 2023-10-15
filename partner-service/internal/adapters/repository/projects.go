package repository

import (
	"errors"
	"fmt"

	"github.com/JIeeiroSst/partner-service/internal/core/domain"
)

func (m *DB) CreateProject(userID string, project domain.Project) error {
	project.ID = snowflakeID()
	project.UserID = userID

	req := m.db.Create(&project)
	if req.RowsAffected == 0 {
		return fmt.Errorf("project not saved: %v", req.Error)
	}

	return nil
}

func (m *DB) ReadProject(id string) (*domain.Project, error) {
	project := &domain.Project{}
	req := m.db.First(&project, "id = ? ", id)
	if req.RowsAffected == 0 {
		return nil, errors.New("project not found")
	}
	return nil, nil
}

func (m *DB) ReadProjects(pagination domain.Pagination) (*domain.Pagination, error) {
	var project []*domain.Project

	m.db.Scopes(paginate(project, "", &pagination, m.db)).Find(&project)
	pagination.Rows = project

	return &pagination, nil
}

func (m *DB) UpdateProject(id string, project domain.Project) error {
	req := m.db.Model(&project).Where("id = ?", id).Updates(project)
	if req.RowsAffected == 0 {
		return errors.New("project not found")
	}
	return nil
}

func (m *DB) DeleteProject(id string) error {
	project := &domain.Project{}
	req := m.db.Where("id = ?", id).Delete(&project)
	if req.RowsAffected == 0 {
		return errors.New("project not found")
	}
	return nil
}
