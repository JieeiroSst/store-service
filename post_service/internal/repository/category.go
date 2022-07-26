package repository

import (
	"github.com/JIeeiroSst/post-service/common"
	"github.com/JIeeiroSst/post-service/model"
	"gorm.io/gorm"
)

type Categories interface {
	Create(category model.Category) error
	Update(id string, category model.Category) error
	Delete(id string) error
	Categories() ([]model.Category, error)
	CategoryById(id string) (*model.Category, error)
}

type CategoryRepo struct {
	db *gorm.DB
}

func NewCategoryRepo(db *gorm.DB) *CategoryRepo {
	return &CategoryRepo{
		db: db,
	}
}

func (r *CategoryRepo) Create(category model.Category) error {
	if err := r.db.Create(&category).Error; err != nil {
		return err
	}
	return nil
}

func (r *CategoryRepo) Update(id string, category model.Category) error {
	if err := r.db.Model(model.Category{}).Where("id = ?", id).Updates(&category).Error; err != nil {
		return err
	}
	return nil
}

func (r *CategoryRepo) Delete(id string) error {
	query := r.db.Delete(model.Category{}, "id = ?", id)
	if query.Error != nil {
		return query.Error
	}
	if query.RowsAffected == 0 {
		return common.NotFound
	}
	return nil
}

func (r *CategoryRepo) Categories() ([]model.Category, error) {
	var categories []model.Category
	query := r.db.Find(&categories)
	if query.Error != nil {
		return nil, query.Error
	}
	if query.RowsAffected == 0 {
		return nil, common.NotFound
	}
	return categories, nil
}

func (r *CategoryRepo) CategoryById(id string) (*model.Category, error) {
	var category model.Category
	query := r.db.Where("id = ?", id).Find(&category)
	if query.Error != nil {
		return nil, query.Error
	}
	if query.RowsAffected == 0 {
		return nil, common.NotFound
	}
	return &category, nil
}
