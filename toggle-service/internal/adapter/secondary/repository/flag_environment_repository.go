package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/JIeeiroSst/toggle-service/internal/domain/model"
	"github.com/JIeeiroSst/toggle-service/internal/domain/port"
)

type flagEnvironmentRepository struct {
	db *gorm.DB
}

func NewFeatureFlagEnvironmentRepository(db *gorm.DB) port.FeatureFlagEnvironmentRepository {
	return &flagEnvironmentRepository{db: db}
}

func (r *flagEnvironmentRepository) GetOrCreate(ctx context.Context, flagID, environmentID uuid.UUID) (*model.FeatureFlagEnvironment, error) {
	var ffe model.FeatureFlagEnvironment
	err := r.db.WithContext(ctx).
		Where("feature_flag_id = ? AND environment_id = ?", flagID, environmentID).
		First(&ffe).Error
	if err == nil {
		return &ffe, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	ffe = model.FeatureFlagEnvironment{FeatureFlagID: flagID, EnvironmentID: environmentID, Enabled: false}
	if err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "feature_flag_id"}, {Name: "environment_id"}}, DoNothing: true}).
		Create(&ffe).Error; err != nil {
		return nil, err
	}
	if ffe.ID == uuid.Nil {
		// Row already existed (race with OnConflict DoNothing) — re-fetch.
		if err := r.db.WithContext(ctx).
			Where("feature_flag_id = ? AND environment_id = ?", flagID, environmentID).
			First(&ffe).Error; err != nil {
			return nil, err
		}
	}
	return &ffe, nil
}

func (r *flagEnvironmentRepository) GetByFlagAndEnv(ctx context.Context, flagID, environmentID uuid.UUID) (*model.FeatureFlagEnvironment, error) {
	var ffe model.FeatureFlagEnvironment
	if err := r.db.WithContext(ctx).
		Where("feature_flag_id = ? AND environment_id = ?", flagID, environmentID).
		First(&ffe).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &ffe, nil
}

func (r *flagEnvironmentRepository) GetByFlagAndEnvWithStrategies(ctx context.Context, flagID, environmentID uuid.UUID) (*model.FeatureFlagEnvironment, error) {
	var ffe model.FeatureFlagEnvironment
	err := r.db.WithContext(ctx).
		Preload("Strategies.Constraints").
		Where("feature_flag_id = ? AND environment_id = ?", flagID, environmentID).
		First(&ffe).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &ffe, nil
}

func (r *flagEnvironmentRepository) SetEnabled(ctx context.Context, id uuid.UUID, enabled bool) error {
	return r.db.WithContext(ctx).
		Model(&model.FeatureFlagEnvironment{}).
		Where("id = ?", id).
		Update("enabled", enabled).Error
}

func (r *flagEnvironmentRepository) ListByEnvironmentWithStrategies(ctx context.Context, projectID, environmentID uuid.UUID) ([]model.FeatureFlagEnvironment, error) {
	var rows []model.FeatureFlagEnvironment
	err := r.db.WithContext(ctx).
		Preload("Strategies.Constraints").
		Preload("FeatureFlag").
		Joins("JOIN feature_flags ON feature_flags.id = feature_flag_environments.feature_flag_id").
		Where("feature_flags.project_id = ? AND feature_flags.archived = false AND feature_flag_environments.environment_id = ?", projectID, environmentID).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}
