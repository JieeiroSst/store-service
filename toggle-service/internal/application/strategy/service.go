package strategy

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"gorm.io/datatypes"

	"github.com/JIeeiroSst/toggle-service/internal/application/apperr"
	"github.com/JIeeiroSst/toggle-service/internal/domain/model"
	"github.com/JIeeiroSst/toggle-service/internal/domain/port"
)

type service struct {
	flags        port.FeatureFlagRepository
	environments port.EnvironmentRepository
	flagEnvs     port.FeatureFlagEnvironmentRepository
	strategies   port.StrategyRepository
	constraints  port.ConstraintRepository
	audit        port.AuditService
}

func NewService(
	flags port.FeatureFlagRepository,
	environments port.EnvironmentRepository,
	flagEnvs port.FeatureFlagEnvironmentRepository,
	strategies port.StrategyRepository,
	constraints port.ConstraintRepository,
	audit port.AuditService,
) port.StrategyService {
	return &service{
		flags: flags, environments: environments, flagEnvs: flagEnvs,
		strategies: strategies, constraints: constraints, audit: audit,
	}
}

func (s *service) resolveFlagEnvironment(ctx context.Context, projectID uuid.UUID, key, environmentName string) (*model.FeatureFlag, *model.Environment, *model.FeatureFlagEnvironment, error) {
	flag, err := s.flags.GetByProjectAndKey(ctx, projectID, key)
	if err != nil {
		return nil, nil, nil, err
	}
	if flag == nil {
		return nil, nil, nil, apperr.ErrNotFound
	}
	env, err := s.environments.GetByName(ctx, environmentName)
	if err != nil {
		return nil, nil, nil, err
	}
	if env == nil {
		return nil, nil, nil, apperr.ErrNotFound
	}
	ffe, err := s.flagEnvs.GetOrCreate(ctx, flag.ID, env.ID)
	if err != nil {
		return nil, nil, nil, err
	}
	return flag, env, ffe, nil
}

func (s *service) Add(ctx context.Context, projectID uuid.UUID, key, environmentName string, in port.StrategyInput, actor string) (*model.ActivationStrategy, error) {
	flag, env, ffe, err := s.resolveFlagEnvironment(ctx, projectID, key, environmentName)
	if err != nil {
		return nil, err
	}

	params, err := json.Marshal(in.Parameters)
	if err != nil {
		return nil, apperr.ErrValidation
	}

	st := &model.ActivationStrategy{
		FeatureFlagEnvironmentID: ffe.ID,
		StrategyType:             in.StrategyType,
		Parameters:               datatypes.JSON(params),
		SortOrder:                in.SortOrder,
	}
	if err := s.strategies.Create(ctx, st); err != nil {
		return nil, err
	}

	for _, c := range in.Constraints {
		if err := s.createConstraint(ctx, st.ID, c); err != nil {
			return nil, err
		}
	}

	st.Constraints, _ = s.constraints.ListByStrategy(ctx, st.ID)
	_ = s.audit.Record(ctx, "activation_strategy", st.ID, model.ActionCreate, &flag.ProjectID, &env.ID, &actor, nil, st)
	return st, nil
}

func (s *service) createConstraint(ctx context.Context, strategyID uuid.UUID, in port.ConstraintInput) error {
	values, err := json.Marshal(in.Values)
	if err != nil {
		return apperr.ErrValidation
	}
	c := &model.Constraint{
		StrategyID:      strategyID,
		ContextField:    in.ContextField,
		Operator:        in.Operator,
		Values:          datatypes.JSON(values),
		CaseInsensitive: in.CaseInsensitive,
	}
	return s.constraints.Create(ctx, c)
}

func (s *service) Update(ctx context.Context, strategyID uuid.UUID, in port.StrategyInput, actor string) (*model.ActivationStrategy, error) {
	st, err := s.strategies.GetByID(ctx, strategyID)
	if err != nil {
		return nil, err
	}
	if st == nil {
		return nil, apperr.ErrNotFound
	}
	before := *st

	params, err := json.Marshal(in.Parameters)
	if err != nil {
		return nil, apperr.ErrValidation
	}
	st.StrategyType = in.StrategyType
	st.Parameters = datatypes.JSON(params)
	st.SortOrder = in.SortOrder
	if err := s.strategies.Update(ctx, st); err != nil {
		return nil, err
	}

	if err := s.constraints.DeleteByStrategy(ctx, st.ID); err != nil {
		return nil, err
	}
	for _, c := range in.Constraints {
		if err := s.createConstraint(ctx, st.ID, c); err != nil {
			return nil, err
		}
	}
	st.Constraints, _ = s.constraints.ListByStrategy(ctx, st.ID)

	_ = s.audit.Record(ctx, "activation_strategy", st.ID, model.ActionUpdate, nil, nil, &actor, before, st)
	return st, nil
}

func (s *service) Delete(ctx context.Context, strategyID uuid.UUID, actor string) error {
	st, err := s.strategies.GetByID(ctx, strategyID)
	if err != nil {
		return err
	}
	if st == nil {
		return apperr.ErrNotFound
	}
	if err := s.strategies.Delete(ctx, strategyID); err != nil {
		return err
	}
	_ = s.audit.Record(ctx, "activation_strategy", strategyID, model.ActionDelete, nil, nil, &actor, st, nil)
	return nil
}

func (s *service) List(ctx context.Context, projectID uuid.UUID, key, environmentName string) ([]model.ActivationStrategy, error) {
	_, _, ffe, err := s.resolveFlagEnvironment(ctx, projectID, key, environmentName)
	if err != nil {
		return nil, err
	}
	return s.strategies.ListByFlagEnvironment(ctx, ffe.ID)
}
