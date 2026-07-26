package repository

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/basket-service/common"
	"github.com/JIeeiroSst/basket-service/internal/domain/model"
	"github.com/JIeeiroSst/basket-service/internal/domain/port"
	"gorm.io/gorm"
)

type basketLineAttributeRepository struct {
	db *gorm.DB
}

func NewBasketLineAttributeRepository(db *gorm.DB) port.BasketLineAttributeRepository {
	return &basketLineAttributeRepository{db: db}
}

func (r *basketLineAttributeRepository) Create(ctx context.Context, attribute *model.BasketLineAttribute) error {
	if err := r.db.WithContext(ctx).Create(attribute).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *basketLineAttributeRepository) GetByID(ctx context.Context, id int) (*model.BasketLineAttribute, error) {
	var attribute model.BasketLineAttribute
	if err := r.db.WithContext(ctx).First(&attribute, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound
		}
		return nil, common.ErrDBFailed
	}
	return &attribute, nil
}

func (r *basketLineAttributeRepository) Update(ctx context.Context, attribute *model.BasketLineAttribute) error {
	if err := r.db.WithContext(ctx).Model(&model.BasketLineAttribute{}).Where("id = ?", attribute.ID).Updates(attribute).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *basketLineAttributeRepository) Delete(ctx context.Context, id int) error {
	if err := r.db.WithContext(ctx).Delete(&model.BasketLineAttribute{}, "id = ?", id).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *basketLineAttributeRepository) List(ctx context.Context) ([]model.BasketLineAttribute, error) {
	var attributes []model.BasketLineAttribute
	if err := r.db.WithContext(ctx).Find(&attributes).Error; err != nil {
		return nil, common.ErrDBFailed
	}
	return attributes, nil
}
