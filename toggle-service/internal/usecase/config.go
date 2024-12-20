package usecase

import (
	"context"
	"time"

	"github.com/JIeeiroSst/utils/time_custom"
)

type Configs interface {
	GetConfigTime(ctx context.Context) map[string]string
	GetTimeCountry(ctx context.Context, country string) time.Time
	GetConfigCountry(ctx context.Context, country string) string
}

type ConfigsUsecase struct {
}

func NewConfigsUsecase() *ConfigsUsecase {
	return &ConfigsUsecase{}
}

func (u *ConfigsUsecase) GetConfigTime(ctx context.Context) map[string]string {
	return time_custom.CountryTz
}

func (u *ConfigsUsecase) GetTimeCountry(ctx context.Context, country string) time.Time {
	return time_custom.TimeInCountry(country)
}

func (u *ConfigsUsecase) GetConfigCountry(ctx context.Context, country string) string {
	return time_custom.CountryTz[country]
}
