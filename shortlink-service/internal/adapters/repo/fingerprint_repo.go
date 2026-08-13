package repo

import (
	"context"
	"time"

	"github.com/JIeeiroSst/shortlink-service/internal/domain"
	"github.com/JIeeiroSst/shortlink-service/internal/ports"
	"gorm.io/gorm"
)

type FingerprintRepo struct{ db *gorm.DB }

func NewFingerprintRepo(db *gorm.DB) *FingerprintRepo { return &FingerprintRepo{db: db} }

func (r *FingerprintRepo) StoreForClick(ctx context.Context, clickID string, data domain.FingerprintData) error {
	hash := domain.GenerateFingerprintHash(data)
	m := &DeviceFingerprintModel{
		ClickID:         clickID,
		FingerprintHash: hash,
		IPAddress:       strPtrOrNil(data.IPAddress),
		UserAgent:       strPtrOrNil(data.UserAgent),
		Timezone:        strPtrOrNil(data.Timezone),
		Language:        strPtrOrNil(data.Language),
		ScreenWidth:     data.ScreenWidth,
		ScreenHeight:    data.ScreenHeight,
		Platform:        strPtrOrNil(data.Platform),
		PlatformVersion: strPtrOrNil(data.PlatformVersion),
	}
	return r.db.WithContext(ctx).Create(m).Error
}

func strPtrOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func (r *FingerprintRepo) CandidateClicks(ctx context.Context, sinceMax time.Time) ([]ports.FingerprintCandidate, error) {
	var rows []struct {
		ClickID                string
		LinkID                 string
		ClickedAt              time.Time
		AttributionWindowHours int
		IPAddress              *string
		UserAgent              *string
		Timezone               *string
		Language               *string
		ScreenWidth            *int
		ScreenHeight           *int
		Platform               *string
		PlatformVersion        *string
	}
	err := r.db.WithContext(ctx).Table("click_events ce").
		Select(`ce.id AS click_id, ce.link_id, ce.clicked_at, l.attribution_window_hours,
		        df.ip_address, df.user_agent, df.timezone, df.language,
		        df.screen_width, df.screen_height, df.platform, df.platform_version`).
		Joins("INNER JOIN device_fingerprints df ON df.click_id = ce.id").
		Joins("INNER JOIN links l ON ce.link_id = l.id").
		Where("ce.clicked_at >= ?", sinceMax).
		Order("ce.clicked_at DESC").
		Limit(1000).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	out := make([]ports.FingerprintCandidate, len(rows))
	for i, row := range rows {
		out[i] = ports.FingerprintCandidate{
			ClickID:                row.ClickID,
			LinkID:                 row.LinkID,
			ClickedAt:              row.ClickedAt,
			AttributionWindowHours: row.AttributionWindowHours,
			Fingerprint: domain.FingerprintData{
				IPAddress:       derefStr(row.IPAddress),
				UserAgent:       derefStr(row.UserAgent),
				Timezone:        derefStr(row.Timezone),
				Language:        derefStr(row.Language),
				ScreenWidth:     row.ScreenWidth,
				ScreenHeight:    row.ScreenHeight,
				Platform:        derefStr(row.Platform),
				PlatformVersion: derefStr(row.PlatformVersion),
			},
		}
	}
	return out, nil
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
