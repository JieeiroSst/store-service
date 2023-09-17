package persistence

import "gorm.io/gorm"

type RewardPointRepositoryImpl struct {
	db *gorm.DB
}

