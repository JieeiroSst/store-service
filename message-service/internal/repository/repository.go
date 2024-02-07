package repository

import "gorm.io/gorm"

type Repositories struct {
	Tracks
}

func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		Tracks: NewTrackRepo(db),
	}
}
