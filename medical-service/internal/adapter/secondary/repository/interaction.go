package repository

import (
	"context"

	"github.com/JIeeiroSst/medical-service/internal/domain/model"
	"github.com/JIeeiroSst/medical-service/internal/domain/port"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type interactionRepository struct {
	db *gorm.DB
}

func NewInteractionRepository(db *gorm.DB) port.InteractionRepository {
	return &interactionRepository{db: db}
}

func (r *interactionRepository) Upsert(ctx context.Context, rule *model.InteractionRule) error {
	db := conn(ctx, r.db)
	err := db.Clauses(clause.OnConflict{
		DoUpdates: clause.AssignmentColumns([]string{"severity", "description", "source", "updated_at"}),
	}).Create(rule).Error
	if err != nil {
		return err
	}

	var stored model.InteractionRule
	err = db.Where("ingredient_a = ? AND ingredient_b = ?", rule.IngredientA, rule.IngredientB).First(&stored).Error
	if err != nil {
		return err
	}
	*rule = stored
	return nil
}

func (r *interactionRepository) List(ctx context.Context, limit, offset int) ([]model.InteractionRule, error) {
	var rules []model.InteractionRule
	err := conn(ctx, r.db).Order("ingredient_a ASC, ingredient_b ASC").Limit(limit).Offset(offset).Find(&rules).Error
	return rules, err
}

func (r *interactionRepository) FindAmong(ctx context.Context, ingredients []string) ([]model.InteractionRule, error) {
	if len(ingredients) < 2 {
		return nil, nil
	}
	var rules []model.InteractionRule
	err := conn(ctx, r.db).
		Where("ingredient_a IN ? AND ingredient_b IN ?", ingredients, ingredients).
		Find(&rules).Error
	return rules, err
}
