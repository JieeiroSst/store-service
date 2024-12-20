package repository

import "gorm.io/gorm"

type Repository struct {
	Countries
	FeatureCountries
}

func NewRepositories(db *gorm.DB) *Repository {
	return &Repository{
		Countries:        NewCountriesRepository(db),
		FeatureCountries: NewFeatureCountriesRepository(db),
	}
}
