package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/JIeeiroSst/toggle-service/internal/domain/model"
	"github.com/JIeeiroSst/toggle-service/internal/domain/port"
)

type environmentRepository struct {
	db *gorm.DB
}

func NewEnvironmentRepository(db *gorm.DB) port.EnvironmentRepository {
	return &environmentRepository{db: db}
}

func (r *environmentRepository) Create(ctx context.Context, e *model.Environment) error {
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *environmentRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Environment, error) {
	var e model.Environment
	if err := r.db.WithContext(ctx).First(&e, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &e, nil
}

func (r *environmentRepository) GetByName(ctx context.Context, name string) (*model.Environment, error) {
	var e model.Environment
	if err := r.db.WithContext(ctx).First(&e, "name = ?", name).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &e, nil
}

func (r *environmentRepository) List(ctx context.Context) ([]model.Environment, error) {
	var envs []model.Environment
	if err := r.db.WithContext(ctx).Order("sort_order asc").Find(&envs).Error; err != nil {
		return nil, err
	}
	return envs, nil
}

func (r *environmentRepository) Update(ctx context.Context, e *model.Environment) error {
	return r.db.WithContext(ctx).Save(e).Error
}

func (r *environmentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Environment{}, "id = ?", id).Error
}
