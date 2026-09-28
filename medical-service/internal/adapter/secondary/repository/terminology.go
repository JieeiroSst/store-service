package repository

import (
	"context"

	"github.com/JIeeiroSst/medical-service/internal/domain/model"
	"github.com/JIeeiroSst/medical-service/internal/domain/port"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type terminologyRepository struct {
	db *gorm.DB
}

func NewTerminologyRepository(db *gorm.DB) port.TerminologyRepository {
	return &terminologyRepository{db: db}
}

func (r *terminologyRepository) Load(ctx context.Context) (model.Terminology, error) {
	aliases, err := r.ListAliases(ctx)
	if err != nil {
		return model.Terminology{}, err
	}
	groups, err := r.ListAllergenGroups(ctx)
	if err != nil {
		return model.Terminology{}, err
	}
	return model.NewTerminology(aliases, groups), nil
}

func (r *terminologyRepository) UpsertAlias(ctx context.Context, a *model.IngredientAlias) error {
	db := conn(ctx, r.db)
	err := db.Clauses(clause.OnConflict{
		DoUpdates: clause.AssignmentColumns([]string{"canonical", "source"}),
	}).Create(a).Error
	if err != nil {
		return err
	}
	var stored model.IngredientAlias
	if err := db.Where("alias = ?", a.Alias).First(&stored).Error; err != nil {
		return err
	}
	*a = stored
	return nil
}

func (r *terminologyRepository) AddAllergenGroup(ctx context.Context, g *model.IngredientAllergenGroup) error {
	db := conn(ctx, r.db)
	if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(g).Error; err != nil {
		return err
	}
	var stored model.IngredientAllergenGroup
	err := db.Where("ingredient = ? AND allergen_group = ?", g.Ingredient, g.AllergenGroup).First(&stored).Error
	if err != nil {
		return err
	}
	*g = stored
	return nil
}

func (r *terminologyRepository) ListAliases(ctx context.Context) ([]model.IngredientAlias, error) {
	var aliases []model.IngredientAlias
	err := conn(ctx, r.db).Order("alias ASC").Find(&aliases).Error
	return aliases, err
}

func (r *terminologyRepository) ListAllergenGroups(ctx context.Context) ([]model.IngredientAllergenGroup, error) {
	var groups []model.IngredientAllergenGroup
	err := conn(ctx, r.db).Order("ingredient ASC, allergen_group ASC").Find(&groups).Error
	return groups, err
}
