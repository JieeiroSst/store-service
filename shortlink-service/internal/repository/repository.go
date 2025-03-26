package repository

import "gorm.io/gorm"

type Repositories struct {
	LinkClicks
	Links
}

func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		LinkClicks: NewLinkClicksRepo(db),
		Links:      NewLinkRepo(db),
	}
}
