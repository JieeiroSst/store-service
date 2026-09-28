package ghn

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/JIeeiroSst/shipping-service/config"
	"github.com/JIeeiroSst/shipping-service/internal/domain/port"
)

const (
	pathProvinces = "/master-data/province"
	pathDistricts = "/master-data/district"
	pathWards     = "/master-data/ward"
)

type cacheEntry struct {
	value   any
	expires time.Time
}

type Locations struct {
	client *Client
	ttl    time.Duration
	now    func() time.Time

	mu    sync.Mutex
	cache map[string]cacheEntry
}

func NewLocations(client *Client, cfg *config.Config) *Locations {
	return &Locations{client: client, ttl: cfg.GHN.LocationCacheTTL(), now: time.Now, cache: map[string]cacheEntry{}}
}

func cached[T any](ctx context.Context, l *Locations, key string, load func(ctx context.Context) (T, error)) (T, error) {
	l.mu.Lock()
	if e, ok := l.cache[key]; ok && l.now().Before(e.expires) {
		l.mu.Unlock()
		return e.value.(T), nil
	}
	l.mu.Unlock()

	v, err := load(ctx)
	if err != nil {
		return v, err
	}
	l.mu.Lock()
	l.cache[key] = cacheEntry{value: v, expires: l.now().Add(l.ttl)}
	l.mu.Unlock()
	return v, nil
}

func (l *Locations) Provinces(ctx context.Context) ([]port.Province, error) {
	return cached(ctx, l, "provinces", func(ctx context.Context) ([]port.Province, error) {
		var data []struct {
			ProvinceID   int    `json:"ProvinceID"`
			ProvinceName string `json:"ProvinceName"`
		}
		if err := l.client.do(ctx, 0, pathProvinces, nil, &data); err != nil {
			return nil, err
		}
		out := make([]port.Province, 0, len(data))
		for _, d := range data {
			out = append(out, port.Province{ID: d.ProvinceID, Name: d.ProvinceName})
		}
		return out, nil
	})
}

func (l *Locations) Districts(ctx context.Context, provinceID int) ([]port.District, error) {
	return cached(ctx, l, "districts:"+strconv.Itoa(provinceID), func(ctx context.Context) ([]port.District, error) {
		var data []struct {
			DistrictID   int    `json:"DistrictID"`
			ProvinceID   int    `json:"ProvinceID"`
			DistrictName string `json:"DistrictName"`
		}
		if err := l.client.do(ctx, 0, pathDistricts, map[string]any{"province_id": provinceID}, &data); err != nil {
			return nil, err
		}
		out := make([]port.District, 0, len(data))
		for _, d := range data {
			out = append(out, port.District{ID: d.DistrictID, ProvinceID: d.ProvinceID, Name: d.DistrictName})
		}
		return out, nil
	})
}

func (l *Locations) Wards(ctx context.Context, districtID int) ([]port.Ward, error) {
	return cached(ctx, l, "wards:"+strconv.Itoa(districtID), func(ctx context.Context) ([]port.Ward, error) {
		var data []struct {
			WardCode   string `json:"WardCode"`
			DistrictID int    `json:"DistrictID"`
			WardName   string `json:"WardName"`
		}
		if err := l.client.do(ctx, 0, pathWards, map[string]any{"district_id": districtID}, &data); err != nil {
			return nil, err
		}
		out := make([]port.Ward, 0, len(data))
		for _, d := range data {
			out = append(out, port.Ward{Code: d.WardCode, DistrictID: d.DistrictID, Name: d.WardName})
		}
		return out, nil
	})
}
