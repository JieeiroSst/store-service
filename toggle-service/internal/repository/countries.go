package repository

import (
	"context"

	"github.com/JIeeiroSst/utils/logger"
	"github.com/JIeerioSst/toggle-service/model"
	"gorm.io/gorm"
)

type Countries interface {
	FindByCountryId(ctx context.Context, countryId int) (*model.Countries, error)
	SaveCountries(ctx context.Context, countries model.Countries) error
	FindByFeatureID(ctx context.Context, featureId int) (*model.Features, error)
	SaveFeatures(ctx context.Context, features model.Features) error
	SaveFeatureCountries(ctx context.Context, featureCountries model.FeatureCountries) error
}

type CountriesRepository struct {
	db *gorm.DB
}

func NewCountriesRepository(db *gorm.DB) *CountriesRepository {
	return &CountriesRepository{
		db: db,
	}
}

func (r *CountriesRepository) FindByCountryId(ctx context.Context, countryId int) (*model.Countries, error) {
	var countries *model.Countries
	if err := r.db.Where("country_id = ?", countryId).Find(&countries).Error; err != nil {
		logger.Error(ctx, "FindByCountryId err %v", err)
		return nil, err
	}
	return countries, nil
}

func (r *CountriesRepository) SaveCountries(ctx context.Context, countries model.Countries) error {
	if err := r.db.Create(&countries).Error; err != nil {
		logger.Error(ctx, "SaveCountries err %v", err)
		return err
	}
	return nil
}

func (r *CountriesRepository) FindByFeatureID(ctx context.Context, featureId int) (*model.Features, error) {
	var features *model.Features
	if err := r.db.Where("feature_id = ?", featureId).Find(&features).Error; err != nil {
		logger.Error(ctx, "FindByCountryId err %v", err)
		return nil, err
	}
	return features, nil
}

func (r *CountriesRepository) SaveFeatures(ctx context.Context, features model.Features) error {
	if err := r.db.Create(&features).Error; err != nil {
		logger.Error(ctx, "SaveFeatures err %v", err)
		return err
	}
	return nil
}

func (r *CountriesRepository) SaveFeatureCountries(ctx context.Context, featureCountries model.FeatureCountries) error {
	if err := r.db.Create(&featureCountries).Error; err != nil {
		return err
	}
	return nil
}
