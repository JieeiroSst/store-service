package repository

import "github.com/JIeeiroSst/partner-service/internal/core/domain"

func (m *DB) CreateProject(userID string, Project domain.Project) error

func (m *DB) ReadProject(id string) (*domain.Project, error)

func (m *DB) ReadProjects() ([]*domain.Project, error)

func (m *DB) UpdateProject(id string, Project domain.Project) error

func (m *DB) DeleteProject(id string) error
