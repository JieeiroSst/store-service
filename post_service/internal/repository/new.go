package repository

import (
	"github.com/JIeeiroSst/post-service/common"
	"github.com/JIeeiroSst/post-service/model"
	"gorm.io/gorm"
)

type News interface {
	Create(new model.New, newCate model.NewCategory) error
	News() ([]model.New, error)
	NewById(id string) (*model.New, error)
	Update(id string, new model.New) error
}

type NewsRepo struct {
	db *gorm.DB
}

func NewsRepository(db *gorm.DB) *NewsRepo {
	return &NewsRepo{
		db: db,
	}
}

func (r *NewsRepo) Create(new model.New, newCate model.NewCategory) error {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := r.db.Create(&new).Error; err != nil {
			return err
		}

		if err := r.db.Create(&newCate).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (r *NewsRepo) News() ([]model.New, error) {
	var news []model.New
	query := r.db.Preload("Categories").Preload("Medias").Find(&news)
	if query.Error != nil {
		return nil, query.Error
	}
	if query.RowsAffected == 0 {
		return nil, common.NotFound
	}
	return news, nil
}

func (r *NewsRepo) NewById(id string) (*model.New, error) {
	var new model.New
	query := r.db.Where("id = ?", id).Preload("Categories").Preload("Medias").Find(&new)
	if query.Error != nil {
		return nil, query.Error
	}
	if query.RowsAffected == 0 {
		return nil, common.NotFound
	}
	return &new, nil
}

func (r *NewsRepo) Update(id string, new model.New) error {
	query := r.db.Where("id = ?", id).Updates(new)
	if query.Error != nil {
		return query.Error
	}
	if query.RowsAffected == 0 {
		return common.NotFound
	}
	return nil
}
