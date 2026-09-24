package repository

import (
	"context"

	"github.com/JIeeiroSst/catalogues-service/internal/domain/model"
	"github.com/JIeeiroSst/catalogues-service/internal/domain/port"
	"gorm.io/gorm"
)

type optionRepo struct{ db *gorm.DB }

func NewOptionRepo(db *gorm.DB) port.OptionRepository { return &optionRepo{db: db} }

func (r *optionRepo) Create(ctx context.Context, o *model.Option) error {
	e := optionToEntity(o)
	if err := r.db.WithContext(ctx).Create(&e).Error; err != nil {
		return writeErr(err)
	}
	o.ID = e.ID
	return nil
}

func (r *optionRepo) Get(ctx context.Context, id int64) (*model.Option, error) {
	var e optionEntity
	if err := r.db.WithContext(ctx).First(&e, id).Error; err != nil {
		return nil, translate(err, nil)
	}
	o := e.toModel()
	return &o, nil
}

func (r *optionRepo) List(ctx context.Context) ([]model.Option, error) {
	var es []optionEntity
	if err := r.db.WithContext(ctx).Order("id").Find(&es).Error; err != nil {
		return nil, err
	}
	out := make([]model.Option, len(es))
	for i, e := range es {
		out[i] = e.toModel()
	}
	return out, nil
}

func (r *optionRepo) Update(ctx context.Context, o *model.Option) error {
	e := optionToEntity(o)
	res := r.db.WithContext(ctx).Model(&optionEntity{ID: o.ID}).Select("*").Omit("id").Updates(&e)
	if res.Error != nil {
		return writeErr(res.Error)
	}
	if res.RowsAffected == 0 {
		return model.ErrNotFound
	}
	return nil
}

func (r *optionRepo) Delete(ctx context.Context, id int64) error {
	return deleteByID[optionEntity](r.db.WithContext(ctx), id)
}
