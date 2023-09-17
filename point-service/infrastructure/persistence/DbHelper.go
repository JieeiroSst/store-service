package persistence

import (
	"github.com/JIeeiroSst/point-service/infrastructure/repository"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type DbHelper struct {
	ConvertedRewardPointRepository repository.ConvertedRewardPointRepository
	RewardDiscountRepository       repository.RewardDiscountRepository
	RewardPointRepository          repository.RewardPointRepository
	db                             *gorm.DB
}

func InitDbHelper(dsn string) (*DbHelper, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	db.AutoMigrate()
	return &DbHelper{
		ConvertedRewardPointRepository: &ConvertedRewardPointRepositoryImpl{db},
		RewardDiscountRepository:       &RewardDiscountRepositoryImpl{db},
		RewardPointRepository:          &RewardPointRepositoryImpl{db},
		db:                             db,
	}, nil
}
