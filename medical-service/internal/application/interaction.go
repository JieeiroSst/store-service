package application

import (
	"context"
	"strings"

	"github.com/JIeeiroSst/medical-service/internal/domain/model"
	"github.com/JIeeiroSst/medical-service/internal/domain/port"
)

type interactionService struct {
	medicines    port.MedicineRepository
	interactions port.InteractionRepository
	terminology  port.TerminologyRepository
}

func NewInteractionService(
	medicines port.MedicineRepository,
	interactions port.InteractionRepository,
	terminology port.TerminologyRepository,
) port.InteractionUsecase {
	return &interactionService{medicines: medicines, interactions: interactions, terminology: terminology}
}

func (s *interactionService) CreateRule(ctx context.Context, in port.CreateInteractionInput) (*model.InteractionRule, error) {
	term, err := s.terminology.Load(ctx)
	if err != nil {
		return nil, err
	}
	a, b := model.IngredientPair(term.Canonical(in.IngredientA), term.Canonical(in.IngredientB))
	switch {
	case a == "" || b == "":
		return nil, invalid("both ingredients are required")
	case a == b:
		return nil, invalid("an interaction needs two different ingredients")
	case !in.Severity.Valid():
		return nil, invalid("severity must be MINOR, MODERATE, MAJOR or CONTRAINDICATED")
	}
	rule := &model.InteractionRule{
		IngredientA: a,
		IngredientB: b,
		Severity:    in.Severity,
		Description: strings.TrimSpace(in.Description),
		Source:      model.SourceLocal,
	}
	if err := s.interactions.Upsert(ctx, rule); err != nil {
		return nil, err
	}
	return rule, nil
}

func (s *interactionService) ListRules(ctx context.Context, in port.ListInput) ([]model.InteractionRule, error) {
	limit, offset := normalizeList(in)
	return s.interactions.List(ctx, limit, offset)
}

func (s *interactionService) Check(ctx context.Context, in port.SafetyCheckInput) (*model.SafetyReport, error) {
	if len(in.MedicineIDs) == 0 {
		return nil, invalid("medicine_ids is required")
	}
	report, err := evaluate(ctx, s.medicines, s.interactions, s.terminology, in.MedicineIDs, in.CurrentMedicineIDs, in.Allergies)
	if err != nil {
		return nil, err
	}
	return report, nil
}

func evaluate(
	ctx context.Context,
	medicines port.MedicineRepository,
	interactions port.InteractionRepository,
	terminology port.TerminologyRepository,
	dispensingIDs, currentIDs []int64,
	allergies []string,
) (*model.SafetyReport, error) {
	dispensing, err := medicines.GetByIDs(ctx, dispensingIDs)
	if err != nil {
		return nil, err
	}
	var current []model.Medicine
	if len(currentIDs) > 0 {
		if current, err = medicines.GetByIDs(ctx, currentIDs); err != nil {
			return nil, err
		}
	}

	term, err := terminology.Load(ctx)
	if err != nil {
		return nil, err
	}
	var ingredients []string
	for _, m := range append(append([]model.Medicine{}, dispensing...), current...) {
		ingredients = append(ingredients, m.Ingredients...)
	}
	rules, err := interactions.FindAmong(ctx, term.CanonicalTerms(ingredients))
	if err != nil {
		return nil, err
	}

	report := model.EvaluateSafety(dispensing, current, allergies, rules, term)
	return &report, nil
}
