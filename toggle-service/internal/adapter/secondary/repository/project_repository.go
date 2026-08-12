package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/JIeeiroSst/toggle-service/internal/domain/model"
	"github.com/JIeeiroSst/toggle-service/internal/domain/port"
)

type projectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) port.ProjectRepository {
	return &projectRepository{db: db}
}

func (r *projectRepository) Create(ctx context.Context, p *model.Project) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *projectRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Project, error) {
	var p model.Project
	if err := r.db.WithContext(ctx).First(&p, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *projectRepository) GetByName(ctx context.Context, name string) (*model.Project, error) {
	var p model.Project
	if err := r.db.WithContext(ctx).First(&p, "name = ?", name).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *projectRepository) List(ctx context.Context) ([]model.Project, error) {
	var projects []model.Project
	if err := r.db.WithContext(ctx).Order("created_at desc").Find(&projects).Error; err != nil {
		return nil, err
	}
	return projects, nil
}

func (r *projectRepository) Update(ctx context.Context, p *model.Project) error {
	return r.db.WithContext(ctx).Save(p).Error
}

func (r *projectRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Project{}, "id = ?", id).Error
}
