package repo

import (
	"context"
	"errors"
	"time"

	"github.com/JIeeiroSst/shortlink-service/internal/domain"
	"github.com/JIeeiroSst/shortlink-service/internal/ports"
	"gorm.io/gorm"
)

type ClickRepo struct{ db *gorm.DB }

func NewClickRepo(db *gorm.DB) *ClickRepo { return &ClickRepo{db: db} }

func clickEventToModel(id string, e *domain.ClickEvent) *ClickEventModel {
	m := &ClickEventModel{
		LinkID:      e.LinkID,
		IPAddress:   e.IPAddress,
		UserAgent:   e.UserAgent,
		DeviceType:  e.DeviceType,
		Platform:    e.Platform,
		CountryCode: e.CountryCode,
		CountryName: e.CountryName,
		Region:      e.Region,
		City:        e.City,
		Latitude:    e.Latitude,
		Longitude:   e.Longitude,
		Timezone:    e.Timezone,
		UTMSource:   e.UTMSource,
		UTMMedium:   e.UTMMedium,
		UTMCampaign: e.UTMCampaign,
		Referrer:    e.Referrer,
		IsBot:       e.IsBot,
		BotReason:   e.BotReason,
	}
	if id != "" {
		m.ID = id
	}
	return m
}

func (r *ClickRepo) Insert(ctx context.Context, event *domain.ClickEvent) error {
	m := clickEventToModel("", event)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	event.ID = m.ID
	event.ClickedAt = m.ClickedAt
	return nil
}

func (r *ClickRepo) InsertWithID(ctx context.Context, id string, event *domain.ClickEvent) error {
	m := clickEventToModel(id, event)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	event.ID = m.ID
	event.ClickedAt = m.ClickedAt
	return nil
}

func (r *ClickRepo) GetByID(ctx context.Context, id string) (*domain.ClickEvent, error) {
	var m ClickEventModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).Take(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &domain.ClickEvent{
		ID: m.ID, LinkID: m.LinkID, ClickedAt: m.ClickedAt, IPAddress: m.IPAddress,
		UserAgent: m.UserAgent, DeviceType: m.DeviceType, Platform: m.Platform,
		CountryCode: m.CountryCode, CountryName: m.CountryName, Region: m.Region,
		City: m.City, Latitude: m.Latitude, Longitude: m.Longitude, Timezone: m.Timezone,
		UTMSource: m.UTMSource, UTMMedium: m.UTMMedium, UTMCampaign: m.UTMCampaign,
		Referrer: m.Referrer, IsBot: m.IsBot, BotReason: m.BotReason,
	}, nil
}

func (r *ClickRepo) scopeAnalytics(ctx context.Context, filter ports.AnalyticsFilter) *gorm.DB {
	days := filter.Days
	if days <= 0 {
		days = 30
	}
	q := r.db.WithContext(ctx).Table("click_events ce").
		Joins("JOIN links l ON ce.link_id = l.id").
		Where("ce.clicked_at >= NOW() - (?::double precision * INTERVAL '1 day')", days).
		Where("ce.is_bot = false")
	if filter.LinkID != nil {
		q = q.Where("ce.link_id = ?", *filter.LinkID)
	}
	if filter.UserID != nil {
		q = q.Where("l.user_id = ?", *filter.UserID)
	}
	return q
}

func (r *ClickRepo) CountTotalAndUnique(ctx context.Context, filter ports.AnalyticsFilter) (int64, int64, error) {
	var row struct {
		Total  int64
		Unique int64
	}
	err := r.scopeAnalytics(ctx, filter).
		Select("COUNT(*) AS total, COUNT(DISTINCT ce.ip_address) AS unique").
		Take(&row).Error
	return row.Total, row.Unique, err
}

func (r *ClickRepo) CountByDate(ctx context.Context, filter ports.AnalyticsFilter) ([]ports.DateCount, error) {
	var rows []struct {
		Date   time.Time
		Clicks int64
	}
	err := r.scopeAnalytics(ctx, filter).
		Select("DATE(ce.clicked_at) AS date, COUNT(*) AS clicks").
		Group("DATE(ce.clicked_at)").Order("date").Scan(&rows).Error
	out := make([]ports.DateCount, len(rows))
	for i, r2 := range rows {
		out[i] = ports.DateCount{Date: r2.Date, Clicks: r2.Clicks}
	}
	return out, err
}

func (r *ClickRepo) CountByCountry(ctx context.Context, filter ports.AnalyticsFilter) ([]ports.CountryCount, error) {
	var rows []struct {
		CountryCode string
		Country     string
		Clicks      int64
	}
	err := r.scopeAnalytics(ctx, filter).
		Select(`COALESCE(ce.country_code, 'Unknown') AS country_code,
		        COALESCE(ce.country_name, ce.country_code, 'Unknown') AS country,
		        COUNT(*) AS clicks`).
		Group("ce.country_code, ce.country_name").Order("clicks DESC").Scan(&rows).Error
	out := make([]ports.CountryCount, len(rows))
	for i, r2 := range rows {
		out[i] = ports.CountryCount{CountryCode: r2.CountryCode, Country: r2.Country, Clicks: r2.Clicks}
	}
	return out, err
}

func (r *ClickRepo) CountByDevice(ctx context.Context, filter ports.AnalyticsFilter) ([]ports.DeviceCount, error) {
	var rows []struct {
		Device string
		Clicks int64
	}
	err := r.scopeAnalytics(ctx, filter).
		Select(`COALESCE(ce.device_type, 'Unknown') AS device, COUNT(*) AS clicks`).
		Group("ce.device_type").Order("clicks DESC").Scan(&rows).Error
	out := make([]ports.DeviceCount, len(rows))
	for i, r2 := range rows {
		out[i] = ports.DeviceCount{Device: r2.Device, Clicks: r2.Clicks}
	}
	return out, err
}

func (r *ClickRepo) CountByPlatform(ctx context.Context, filter ports.AnalyticsFilter) ([]ports.PlatformCount, error) {
	var rows []struct {
		Platform string
		Clicks   int64
	}
	err := r.scopeAnalytics(ctx, filter).
		Select(`COALESCE(ce.platform, 'Unknown') AS platform, COUNT(*) AS clicks`).
		Group("ce.platform").Order("clicks DESC").Scan(&rows).Error
	out := make([]ports.PlatformCount, len(rows))
	for i, r2 := range rows {
		out[i] = ports.PlatformCount{Platform: r2.Platform, Clicks: r2.Clicks}
	}
	return out, err
}

func (r *ClickRepo) TopLinks(ctx context.Context, filter ports.AnalyticsFilter, limit int) ([]ports.TopLink, error) {
	days := filter.Days
	if days <= 0 {
		days = 30
	}
	q := r.db.WithContext(ctx).Table("links l").
		Joins(`LEFT JOIN click_events ce ON l.id = ce.link_id
		       AND ce.clicked_at >= NOW() - (?::double precision * INTERVAL '1 day')
		       AND ce.is_bot = false`, days)
	if filter.UserID != nil {
		q = q.Where("l.user_id = ?", *filter.UserID)
	}
	var rows []struct {
		ID           string
		ShortCode    string
		Title        *string
		OriginalURL  string
		TotalClicks  int64
		UniqueClicks int64
	}
	err := q.Select(`l.id, l.short_code, l.title, l.original_url,
	                  COUNT(ce.id) AS total_clicks, COUNT(DISTINCT ce.ip_address) AS unique_clicks`).
		Group("l.id").Order("total_clicks DESC").Limit(limit).Scan(&rows).Error
	out := make([]ports.TopLink, len(rows))
	for i, r2 := range rows {
		out[i] = ports.TopLink{
			ID: r2.ID, ShortCode: r2.ShortCode, Title: r2.Title, OriginalURL: r2.OriginalURL,
			TotalClicks: r2.TotalClicks, UniqueClicks: r2.UniqueClicks,
		}
	}
	return out, err
}
