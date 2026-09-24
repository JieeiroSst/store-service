package repository

import (
	"context"

	"github.com/JIeeiroSst/catalogues-service/internal/domain/model"
	"github.com/JIeeiroSst/catalogues-service/internal/domain/port"
	"gorm.io/gorm"
)

type categoryRepo struct{ db *gorm.DB }

func NewCategoryRepo(db *gorm.DB) port.CategoryRepository { return &categoryRepo{db: db} }

func (r *categoryRepo) Create(ctx context.Context, c *model.Category) error {
	e := categoryToEntity(c)
	if err := r.db.WithContext(ctx).Omit("Parent").Create(&e).Error; err != nil {
		return writeErr(err)
	}
	c.ID = e.ID
	return nil
}

func (r *categoryRepo) Get(ctx context.Context, id int64) (*model.Category, error) {
	var e categoryEntity
	if err := r.db.WithContext(ctx).First(&e, id).Error; err != nil {
		return nil, translate(err, nil)
	}
	c := e.toModel()
	return &c, nil
}

func (r *categoryRepo) List(ctx context.Context) ([]model.Category, error) {
	var es []categoryEntity
	if err := r.db.WithContext(ctx).Order("id").Find(&es).Error; err != nil {
		return nil, err
	}
	out := make([]model.Category, len(es))
	for i, e := range es {
		out[i] = e.toModel()
	}
	return out, nil
}

func (r *categoryRepo) Update(ctx context.Context, c *model.Category) error {
	e := categoryToEntity(c)
	res := r.db.WithContext(ctx).Model(&categoryEntity{ID: c.ID}).Select("*").Omit("id", "Parent").Updates(&e)
	if res.Error != nil {
		return writeErr(res.Error)
	}
	if res.RowsAffected == 0 {
		return model.ErrNotFound
	}
	return nil
}

func (r *categoryRepo) Delete(ctx context.Context, id int64) error {
	return deleteByID[categoryEntity](r.db.WithContext(ctx), id)
}
