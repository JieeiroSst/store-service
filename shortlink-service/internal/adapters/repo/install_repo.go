package repo

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/shortlink-service/internal/domain"
	"gorm.io/gorm"
)

type InstallRepo struct{ db *gorm.DB }

func NewInstallRepo(db *gorm.DB) *InstallRepo { return &InstallRepo{db: db} }

func (r *InstallRepo) Insert(ctx context.Context, event *domain.InstallEvent) (string, error) {
	m := &InstallEventModel{
		LinkID:                 event.LinkID,
		ClickID:                event.ClickID,
		FingerprintHash:        event.FingerprintHash,
		ConfidenceScore:        event.ConfidenceScore,
		AttributionMethod:      event.AttributionMethod,
		MatchedFactors:         event.MatchedFactors,
		FirstOpenAt:            event.FirstOpenAt,
		DeepLinkRetrieved:      event.DeepLinkRetrieved,
		DeepLinkData:           mapToJSON(event.DeepLinkData),
		AttributionWindowHours: event.AttributionWindowHours,
		IPAddress:              event.IPAddress,
		UserAgent:              event.UserAgent,
		Timezone:               event.Timezone,
		Language:               event.Language,
		ScreenWidth:            event.ScreenWidth,
		ScreenHeight:           event.ScreenHeight,
		Platform:               event.Platform,
		PlatformVersion:        event.PlatformVersion,
		DeviceID:               event.DeviceID,
		SDKName:                event.SDKName,
		SDKVersion:             event.SDKVersion,
	}
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return "", err
	}
	return m.ID, nil
}

func (r *InstallRepo) modelToDomain(m *InstallEventModel) *domain.InstallEvent {
	return &domain.InstallEvent{
		ID: m.ID, LinkID: m.LinkID, ClickID: m.ClickID, FingerprintHash: m.FingerprintHash,
		ConfidenceScore: m.ConfidenceScore, AttributionMethod: m.AttributionMethod,
		MatchedFactors: []string(m.MatchedFactors), InstalledAt: m.InstalledAt, FirstOpenAt: m.FirstOpenAt,
		DeepLinkRetrieved: m.DeepLinkRetrieved, DeepLinkData: jsonToMap(m.DeepLinkData),
		AttributionWindowHours: m.AttributionWindowHours, IPAddress: m.IPAddress, UserAgent: m.UserAgent,
		Timezone: m.Timezone, Language: m.Language, ScreenWidth: m.ScreenWidth, ScreenHeight: m.ScreenHeight,
		Platform: m.Platform, PlatformVersion: m.PlatformVersion, DeviceID: m.DeviceID,
		SDKName: m.SDKName, SDKVersion: m.SDKVersion, CreatedAt: m.CreatedAt,
	}
}

func (r *InstallRepo) GetByID(ctx context.Context, id string) (*domain.InstallEvent, error) {
	var m InstallEventModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).Take(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return r.modelToDomain(&m), nil
}

func (r *InstallRepo) SetDeepLinkData(ctx context.Context, id string, data map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&InstallEventModel{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"deep_link_data":      mapToJSON(data),
			"deep_link_retrieved": true,
		}).Error
}

func (r *InstallRepo) LatestByFingerprintHash(ctx context.Context, hash string) (*domain.InstallEvent, error) {
	var m InstallEventModel
	err := r.db.WithContext(ctx).Where("fingerprint_hash = ?", hash).
		Order("installed_at DESC").Limit(1).Take(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return r.modelToDomain(&m), nil
}
