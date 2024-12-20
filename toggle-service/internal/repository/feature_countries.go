package repository

import (
	"context"

	"github.com/JIeerioSst/toggle-service/model"
	"gorm.io/gorm"
)

type FeatureCountries interface {
	SaveUsers(ctx context.Context, user model.Users) error
	FindByUserID(ctx context.Context, userID int) (*model.Users, error)
	FindFeatureCountries(ctx context.Context, countryId int) (*model.FeatureCountries, error)
}

type FeatureCountriesRepository struct {
	db *gorm.DB
}

func NewFeatureCountriesRepository(db *gorm.DB) *FeatureCountriesRepository {
	return &FeatureCountriesRepository{
		db: db,
	}
}

func (r *FeatureCountriesRepository) FindByUserID(ctx context.Context, userID int) (*model.Users, error) {
	var user *model.Users
	if err := r.db.Where("user_id = ?", userID).Find(&user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (r *FeatureCountriesRepository) SaveUsers(ctx context.Context, user model.Users) error {
	model, err := r.FindByUserID(ctx, user.UserId)
	if err != nil {
		return err
	}
	if model.UserId == 0 {
		if err := r.db.Create(&user).Error; err != nil {
			return err
		}
	}
	if err := r.db.Save(&user).Error; err != nil {
		return err
	}
	return nil
}

func (r *FeatureCountriesRepository) FindFeatureCountries(ctx context.Context, countryId int) (*model.FeatureCountries, error) {
	var featureCountries *model.FeatureCountries
	if err := r.db.Where("country_id = ?", countryId).Preload("Features").
		Preload("Countries").Find(&featureCountries).Error; err != nil {
		return nil, err
	}
	return featureCountries, nil
}
