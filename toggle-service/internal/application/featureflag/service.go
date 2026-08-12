package featureflag

import (
	"context"

	"github.com/google/uuid"

	"github.com/JIeeiroSst/toggle-service/internal/application/apperr"
	"github.com/JIeeiroSst/toggle-service/internal/domain/model"
	"github.com/JIeeiroSst/toggle-service/internal/domain/port"
)

type service struct {
	flags        port.FeatureFlagRepository
	flagEnvs     port.FeatureFlagEnvironmentRepository
	environments port.EnvironmentRepository
	audit        port.AuditService
}

func NewService(
	flags port.FeatureFlagRepository,
	flagEnvs port.FeatureFlagEnvironmentRepository,
	environments port.EnvironmentRepository,
	audit port.AuditService,
) port.FeatureFlagService {
	return &service{flags: flags, flagEnvs: flagEnvs, environments: environments, audit: audit}
}

func (s *service) Create(ctx context.Context, in port.CreateFlagInput, actor uuid.UUID) (*model.FeatureFlag, error) {
	existing, err := s.flags.GetByProjectAndKey(ctx, in.ProjectID, in.Key)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, apperr.ErrConflict
	}

	flagType := in.Type
	if flagType == "" {
		flagType = model.FlagTypeRelease
	}
	f := &model.FeatureFlag{
		ProjectID:   in.ProjectID,
		Key:         in.Key,
		Name:        in.Name,
		Description: in.Description,
		Type:        flagType,
		CreatedBy:   actor,
	}
	if err := s.flags.Create(ctx, f); err != nil {
		return nil, err
	}
	
	envs, err := s.environments.List(ctx)
	if err == nil {
		for _, e := range envs {
			_, _ = s.flagEnvs.GetOrCreate(ctx, f.ID, e.ID)
		}
	}

	_ = s.audit.Record(ctx, "feature_flag", f.ID, model.ActionCreate, &f.ProjectID, nil, &actor, nil, f)
	return f, nil
}

func (s *service) Get(ctx context.Context, projectID uuid.UUID, key string) (*model.FeatureFlag, error) {
	f, err := s.flags.GetByProjectAndKey(ctx, projectID, key)
	if err != nil {
		return nil, err
	}
	if f == nil {
		return nil, apperr.ErrNotFound
	}
	return f, nil
}

func (s *service) List(ctx context.Context, projectID uuid.UUID) ([]model.FeatureFlag, error) {
	return s.flags.ListByProject(ctx, projectID)
}

func (s *service) Update(ctx context.Context, projectID uuid.UUID, key string, in port.UpdateFlagInput, actor uuid.UUID) (*model.FeatureFlag, error) {
	f, err := s.Get(ctx, projectID, key)
	if err != nil {
		return nil, err
	}
	before := *f
	if in.Name != "" {
		f.Name = in.Name
	}
	f.Description = in.Description
	if in.Type != "" {
		f.Type = in.Type
	}
	if err := s.flags.Update(ctx, f); err != nil {
		return nil, err
	}
	_ = s.audit.Record(ctx, "feature_flag", f.ID, model.ActionUpdate, &f.ProjectID, nil, &actor, before, f)
	return f, nil
}

func (s *service) Archive(ctx context.Context, projectID uuid.UUID, key string, actor uuid.UUID) error {
	f, err := s.Get(ctx, projectID, key)
	if err != nil {
		return err
	}
	before := *f
	f.Archived = true
	if err := s.flags.Update(ctx, f); err != nil {
		return err
	}
	_ = s.audit.Record(ctx, "feature_flag", f.ID, model.ActionDelete, &f.ProjectID, nil, &actor, before, f)
	return nil
}

func (s *service) Toggle(ctx context.Context, projectID uuid.UUID, key, environmentName string, enabled bool, actor uuid.UUID) error {
	f, err := s.Get(ctx, projectID, key)
	if err != nil {
		return err
	}
	env, err := s.environments.GetByName(ctx, environmentName)
	if err != nil {
		return err
	}
	if env == nil {
		return apperr.ErrNotFound
	}

	ffe, err := s.flagEnvs.GetOrCreate(ctx, f.ID, env.ID)
	if err != nil {
		return err
	}
	if err := s.flagEnvs.SetEnabled(ctx, ffe.ID, enabled); err != nil {
		return err
	}

	action := model.ActionToggleOff
	if enabled {
		action = model.ActionToggleOn
	}
	_ = s.audit.Record(ctx, "feature_flag_environment", ffe.ID, action, &f.ProjectID, &env.ID, &actor,
		map[string]bool{"enabled": ffe.Enabled}, map[string]bool{"enabled": enabled})
	return nil
}
