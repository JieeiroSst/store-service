package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/JIeeiroSst/toggle-service/internal/domain/model"
	"github.com/JIeeiroSst/toggle-service/internal/domain/port"
)

type metricsRepository struct {
	db *gorm.DB
}

func NewMetricsRepository(db *gorm.DB) port.MetricsRepository {
	return &metricsRepository{db: db}
}

func (r *metricsRepository) IncrementCounts(ctx context.Context, flagID, environmentID uuid.UUID, appName string, windowStart, windowStop time.Time, yes, no int64) error {
	return r.db.WithContext(ctx).Exec(`
		INSERT INTO feature_usage_metrics
			(id, feature_flag_id, environment_id, app_name, yes_count, no_count, window_start, window_stop, created_at)
		VALUES (gen_random_uuid(), ?, ?, ?, ?, ?, ?, ?, now())
		ON CONFLICT (feature_flag_id, environment_id, app_name, window_start)
		DO UPDATE SET
			yes_count = feature_usage_metrics.yes_count + EXCLUDED.yes_count,
			no_count = feature_usage_metrics.no_count + EXCLUDED.no_count,
			window_stop = EXCLUDED.window_stop
	`, flagID, environmentID, appName, yes, no, windowStart, windowStop).Error
}

func (r *metricsRepository) ListByFlag(ctx context.Context, flagID uuid.UUID) ([]model.FeatureUsageMetric, error) {
	var metrics []model.FeatureUsageMetric
	if err := r.db.WithContext(ctx).
		Where("feature_flag_id = ?", flagID).
		Order("window_start desc").
		Find(&metrics).Error; err != nil {
		return nil, err
	}
	return metrics, nil
}
