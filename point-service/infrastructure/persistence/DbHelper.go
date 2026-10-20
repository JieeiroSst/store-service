package persistence

import (
	"github.com/JIeeiroSst/point-service/domain/entity"
	"github.com/JIeeiroSst/point-service/infrastructure/repository"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type DbHelper struct {
	ConvertedRewardPointRepository repository.ConvertedRewardPointRepository
	RewardDiscountRepository       repository.RewardDiscountRepository
	RewardPointRepository          repository.RewardPointRepository
	PointLedgerRepository          repository.PointLedgerRepository
	db                             *gorm.DB
}

func InitDbHelper(dsn string) (*DbHelper, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	db.AutoMigrate(&entity.ConvertedRewardPoint{}, &entity.RewardDiscount{}, &entity.RewardPoint{})
	if err := db.AutoMigrate(&entity.PointAccount{}, &entity.PointTransaction{}); err != nil {
		return nil, err
	}
	return &DbHelper{
		ConvertedRewardPointRepository: &ConvertedRewardPointRepositoryImpl{db},
		RewardDiscountRepository:       &RewardDiscountRepositoryImpl{db},
		RewardPointRepository:          &RewardPointRepositoryImpl{db},
		PointLedgerRepository:          &PointLedgerRepositoryImpl{db},
		db:                             db,
	}, nil
}
