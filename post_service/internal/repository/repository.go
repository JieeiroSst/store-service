package repository

import "gorm.io/gorm"

type Repositories struct {
	Medias
	Categories
	News
}

func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		Medias:     NewMediaRepo(db),
		Categories: NewCategoryRepo(db),
		News:       NewsRepository(db),
	}
}
