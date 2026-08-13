package app

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/shortlink-service/internal/adapters/repo"
	"github.com/JIeeiroSst/shortlink-service/internal/ports"
)

type AnalyticsService struct {
	links  ports.LinkRepository
	clicks ports.ClickEventRepository
}

func NewAnalyticsService(links ports.LinkRepository, clicks ports.ClickEventRepository) *AnalyticsService {
	return &AnalyticsService{links, clicks}
}

type AnalyticsOverview struct {
	TotalClicks      int64
	UniqueClicks     int64
	ClicksByDate     []ports.DateCount
	ClicksByCountry  []ports.CountryCount
	ClicksByDevice   []ports.DeviceCount
	ClicksByPlatform []ports.PlatformCount
	TopLinks         []ports.TopLink
}

func (s *AnalyticsService) Overview(ctx context.Context, userID *string, days int) (*AnalyticsOverview, error) {
	filter := ports.AnalyticsFilter{UserID: userID, Days: days}

	total, unique, err := s.clicks.CountTotalAndUnique(ctx, filter)
	if err != nil {
		return nil, err
	}
	byDate, err := s.clicks.CountByDate(ctx, filter)
	if err != nil {
		return nil, err
	}
	byCountry, err := s.clicks.CountByCountry(ctx, filter)
	if err != nil {
		return nil, err
	}
	byDevice, err := s.clicks.CountByDevice(ctx, filter)
	if err != nil {
		return nil, err
	}
	byPlatform, err := s.clicks.CountByPlatform(ctx, filter)
	if err != nil {
		return nil, err
	}
	topLinks, err := s.clicks.TopLinks(ctx, filter, 10)
	if err != nil {
		return nil, err
	}

	return &AnalyticsOverview{
		TotalClicks: total, UniqueClicks: unique, ClicksByDate: byDate, ClicksByCountry: byCountry,
		ClicksByDevice: byDevice, ClicksByPlatform: byPlatform, TopLinks: topLinks,
	}, nil
}

type LinkAnalytics struct {
	TotalClicks      int64
	UniqueClicks     int64
	ClicksByDate     []ports.DateCount
	ClicksByCountry  []ports.CountryCount
	ClicksByDevice   []ports.DeviceCount
	ClicksByPlatform []ports.PlatformCount
}

func (s *AnalyticsService) ForLink(ctx context.Context, linkID string, userID *string, days int) (*LinkAnalytics, error) {
	_, err := s.links.GetByID(ctx, linkID, ports.LinkFilter{UserID: userID})
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrLinkNotFound
		}
		return nil, err
	}

	filter := ports.AnalyticsFilter{LinkID: &linkID, Days: days}
	total, unique, err := s.clicks.CountTotalAndUnique(ctx, filter)
	if err != nil {
		return nil, err
	}
	byDate, err := s.clicks.CountByDate(ctx, filter)
	if err != nil {
		return nil, err
	}
	byCountry, err := s.clicks.CountByCountry(ctx, filter)
	if err != nil {
		return nil, err
	}
	byDevice, err := s.clicks.CountByDevice(ctx, filter)
	if err != nil {
		return nil, err
	}
	byPlatform, err := s.clicks.CountByPlatform(ctx, filter)
	if err != nil {
		return nil, err
	}

	return &LinkAnalytics{
		TotalClicks: total, UniqueClicks: unique, ClicksByDate: byDate, ClicksByCountry: byCountry,
		ClicksByDevice: byDevice, ClicksByPlatform: byPlatform,
	}, nil
}
