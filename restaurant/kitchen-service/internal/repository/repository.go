package repository

import "gorm.io/gorm"

type Repository struct {
	Categories
	Foods
	Kitchens
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{
		Categories: NewCategories(db),
		Foods:      NewFoodRepository(db),
		Kitchens:   NewKitchenRepository(db),
	}
}
