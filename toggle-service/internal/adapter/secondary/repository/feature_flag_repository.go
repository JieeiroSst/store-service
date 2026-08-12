package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/JIeeiroSst/toggle-service/internal/domain/model"
	"github.com/JIeeiroSst/toggle-service/internal/domain/port"
)

type featureFlagRepository struct {
	db *gorm.DB
}

func NewFeatureFlagRepository(db *gorm.DB) port.FeatureFlagRepository {
	return &featureFlagRepository{db: db}
}

func (r *featureFlagRepository) Create(ctx context.Context, f *model.FeatureFlag) error {
	return r.db.WithContext(ctx).Create(f).Error
}

func (r *featureFlagRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.FeatureFlag, error) {
	var f model.FeatureFlag
	if err := r.db.WithContext(ctx).First(&f, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &f, nil
}

func (r *featureFlagRepository) GetByProjectAndKey(ctx context.Context, projectID uuid.UUID, key string) (*model.FeatureFlag, error) {
	var f model.FeatureFlag
	if err := r.db.WithContext(ctx).First(&f, "project_id = ? AND key = ?", projectID, key).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &f, nil
}

func (r *featureFlagRepository) ListByProject(ctx context.Context, projectID uuid.UUID) ([]model.FeatureFlag, error) {
	var flags []model.FeatureFlag
	if err := r.db.WithContext(ctx).
		Where("project_id = ? AND archived = false", projectID).
		Order("created_at desc").
		Find(&flags).Error; err != nil {
		return nil, err
	}
	return flags, nil
}

func (r *featureFlagRepository) Update(ctx context.Context, f *model.FeatureFlag) error {
	return r.db.WithContext(ctx).Save(f).Error
}

func (r *featureFlagRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.FeatureFlag{}, "id = ?", id).Error
}
