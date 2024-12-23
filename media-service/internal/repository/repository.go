package repository

import (
	"github.com/olivere/elastic/v7"
	"gorm.io/gorm"
)

type Repository struct {
	Subscription
	View
	Video
}

func NewRepositories(db *gorm.DB, elastic *elastic.Client) *Repository {
	return &Repository{
		Subscription: NewSubscriptionRepository(db),
		View:         NewViewRepository(db),
		Video:        NewVideoRepository(db, elastic),
	}
}
