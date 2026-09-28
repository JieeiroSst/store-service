package application

import (
	"context"

	"github.com/JIeeiroSst/medical-service/internal/domain/model"
	"github.com/JIeeiroSst/medical-service/internal/domain/port"
)

type terminologyService struct {
	terminology port.TerminologyRepository
}

func NewTerminologyService(terminology port.TerminologyRepository) port.TerminologyUsecase {
	return &terminologyService{terminology: terminology}
}

func (s *terminologyService) AddAlias(ctx context.Context, alias, canonical string) (*model.IngredientAlias, error) {
	alias, canonical = model.NormalizeTerm(alias), model.NormalizeTerm(canonical)
	switch {
	case alias == "" || canonical == "":
		return nil, invalid("alias and canonical are required")
	case alias == canonical:
		return nil, invalid("alias and canonical must differ")
	}

	term, err := s.terminology.Load(ctx)
	if err != nil {
		return nil, err
	}
	if term.IsAlias(canonical) {
		return nil, invalid("%q is itself an alias of %q; point %q at %q instead",
			canonical, term.Canonical(canonical), alias, term.Canonical(canonical))
	}
	if term.IsCanonicalTarget(alias) {
		return nil, invalid("%q is already the canonical name other aliases point to", alias)
	}

	a := &model.IngredientAlias{Alias: alias, Canonical: canonical, Source: model.SourceLocal}
	if err := s.terminology.UpsertAlias(ctx, a); err != nil {
		return nil, err
	}
	return a, nil
}

func (s *terminologyService) AddAllergenGroup(ctx context.Context, ingredient, group string) (*model.IngredientAllergenGroup, error) {
	term, err := s.terminology.Load(ctx)
	if err != nil {
		return nil, err
	}
	ingredient, group = term.Canonical(ingredient), term.Canonical(group)
	switch {
	case ingredient == "" || group == "":
		return nil, invalid("ingredient and allergen_group are required")
	case ingredient == group:
		return nil, invalid("an ingredient cannot be its own allergen group")
	}

	g := &model.IngredientAllergenGroup{Ingredient: ingredient, AllergenGroup: group, Source: model.SourceLocal}
	if err := s.terminology.AddAllergenGroup(ctx, g); err != nil {
		return nil, err
	}
	return g, nil
}

func (s *terminologyService) List(ctx context.Context) (*port.TerminologyList, error) {
	aliases, err := s.terminology.ListAliases(ctx)
	if err != nil {
		return nil, err
	}
	groups, err := s.terminology.ListAllergenGroups(ctx)
	if err != nil {
		return nil, err
	}
	return &port.TerminologyList{Aliases: aliases, AllergenGroups: groups}, nil
}
