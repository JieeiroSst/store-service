package repository

import (
	"context"

	"github.com/JIeeiroSst/catalogues-service/internal/domain/model"
	"github.com/JIeeiroSst/catalogues-service/internal/domain/port"
	"gorm.io/gorm"
)

type productClassRepo struct{ db *gorm.DB }

func NewProductClassRepo(db *gorm.DB) port.ProductClassRepository {
	return &productClassRepo{db: db}
}

func classToEntity(c *model.ProductClass) productClassEntity {
	return productClassEntity{
		ID: c.ID, Name: c.Name, Slug: c.Slug,
		RequiresShipping: c.RequiresShipping, TrackStock: c.TrackStock,
	}
}

func setClassOptions(tx *gorm.DB, classID int64, optionIDs []int64) error {
	if err := tx.Where("product_class_id = ?", classID).Delete(&productClassOptionEntity{}).Error; err != nil {
		return err
	}
	if len(optionIDs) == 0 {
		return nil
	}
	rows := make([]productClassOptionEntity, len(optionIDs))
	for i, id := range optionIDs {
		rows[i] = productClassOptionEntity{ProductClassID: classID, OptionID: id}
	}
	return tx.Omit("ProductClass", "Option").Create(&rows).Error
}

func (r *productClassRepo) Create(ctx context.Context, c *model.ProductClass) error {
	e := classToEntity(c)
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&e).Error; err != nil {
			return err
		}
		return setClassOptions(tx, e.ID, c.OptionIDs)
	})
	if err != nil {
		return writeErr(err)
	}
	c.ID = e.ID
	return nil
}

// hydrate fills OptionIDs for every class in one query.
func (r *productClassRepo) hydrate(ctx context.Context, es []productClassEntity) ([]model.ProductClass, error) {
	out := make([]model.ProductClass, len(es))
	idx := make(map[int64]int, len(es))
	ids := make([]int64, len(es))
	for i, e := range es {
		out[i] = model.ProductClass{
			ID: e.ID, Name: e.Name, Slug: e.Slug,
			RequiresShipping: e.RequiresShipping, TrackStock: e.TrackStock,
			OptionIDs: []int64{},
		}
		idx[e.ID], ids[i] = i, e.ID
	}
	if len(ids) == 0 {
		return out, nil
	}
	var links []productClassOptionEntity
	if err := r.db.WithContext(ctx).Where("product_class_id IN ?", ids).Order("option_id").Find(&links).Error; err != nil {
		return nil, err
	}
	for _, l := range links {
		i := idx[l.ProductClassID]
		out[i].OptionIDs = append(out[i].OptionIDs, l.OptionID)
	}
	return out, nil
}

func (r *productClassRepo) Get(ctx context.Context, id int64) (*model.ProductClass, error) {
	var e productClassEntity
	if err := r.db.WithContext(ctx).First(&e, id).Error; err != nil {
		return nil, translate(err, nil)
	}
	cs, err := r.hydrate(ctx, []productClassEntity{e})
	if err != nil {
		return nil, err
	}
	return &cs[0], nil
}

func (r *productClassRepo) List(ctx context.Context) ([]model.ProductClass, error) {
	var es []productClassEntity
	if err := r.db.WithContext(ctx).Order("id").Find(&es).Error; err != nil {
		return nil, err
	}
	return r.hydrate(ctx, es)
}

func (r *productClassRepo) Update(ctx context.Context, c *model.ProductClass) error {
	e := classToEntity(c)
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&productClassEntity{ID: c.ID}).Select("*").Omit("id").Updates(&e)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return model.ErrNotFound
		}
		return setClassOptions(tx, c.ID, c.OptionIDs)
	})
	return writeErr(err)
}

func (r *productClassRepo) Delete(ctx context.Context, id int64) error {
	return deleteErr(r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("product_class_id = ?", id).Delete(&productClassOptionEntity{}).Error; err != nil {
			return err
		}
		return deleteByID[productClassEntity](tx, id)
	}))
}
