package client

import (
	"context"

	"github.com/JIeeiroSst/toggle-service/internal/application/apperr"
	"github.com/JIeeiroSst/toggle-service/internal/application/evaluation"
	"github.com/JIeeiroSst/toggle-service/internal/domain/model"
	"github.com/JIeeiroSst/toggle-service/internal/domain/port"
)

type service struct {
	flags    port.FeatureFlagRepository
	flagEnvs port.FeatureFlagEnvironmentRepository
	metrics  port.MetricsRepository
	engine   *evaluation.Engine
}

func NewService(
	flags port.FeatureFlagRepository,
	flagEnvs port.FeatureFlagEnvironmentRepository,
	metrics port.MetricsRepository,
	engine *evaluation.Engine,
) port.ClientService {
	return &service{flags: flags, flagEnvs: flagEnvs, metrics: metrics, engine: engine}
}

func (s *service) GetFeatures(ctx context.Context, tok *model.APIToken) (*port.ClientFeaturesResponse, error) {
	if tok.ProjectID == nil || tok.EnvironmentID == nil {
		return nil, apperr.ErrForbidden
	}

	rows, err := s.flagEnvs.ListByEnvironmentWithStrategies(ctx, *tok.ProjectID, *tok.EnvironmentID)
	if err != nil {
		return nil, err
	}

	features := make([]port.ClientFeature, 0, len(rows))
	for _, row := range rows {
		if row.FeatureFlag == nil {
			continue
		}
		features = append(features, port.ClientFeature{
			Name:        row.FeatureFlag.Key,
			Description: row.FeatureFlag.Description,
			Type:        row.FeatureFlag.Type,
			Enabled:     row.Enabled,
			Strategies:  row.Strategies,
		})
	}

	return &port.ClientFeaturesResponse{Version: 1, Features: features}, nil
}

func (s *service) IngestMetrics(ctx context.Context, tok *model.APIToken, payload port.MetricsPayload) error {
	if tok.ProjectID == nil || tok.EnvironmentID == nil {
		return apperr.ErrForbidden
	}

	for flagKey, data := range payload.Bucket.Toggles {
		flag, err := s.flags.GetByProjectAndKey(ctx, *tok.ProjectID, flagKey)
		if err != nil {
			return err
		}
		if flag == nil {
			continue
		}
		if err := s.metrics.IncrementCounts(ctx, flag.ID, *tok.EnvironmentID, payload.AppName,
			payload.Bucket.Start, payload.Bucket.Stop, data.Yes, data.No); err != nil {
			return err
		}
	}
	return nil
}

func (s *service) Evaluate(ctx context.Context, tok *model.APIToken, flagKey string, evalCtx model.EvaluationContext) (bool, error) {
	if tok.ProjectID == nil || tok.EnvironmentID == nil {
		return false, apperr.ErrForbidden
	}

	flag, err := s.flags.GetByProjectAndKey(ctx, *tok.ProjectID, flagKey)
	if err != nil {
		return false, err
	}
	if flag == nil {
		return false, apperr.ErrNotFound
	}

	ffe, err := s.flagEnvs.GetByFlagAndEnvWithStrategies(ctx, flag.ID, *tok.EnvironmentID)
	if err != nil {
		return false, err
	}
	if ffe == nil {
		return false, nil
	}

	return s.engine.Evaluate(flagKey, ffe, ffe.Strategies, evalCtx), nil
}
