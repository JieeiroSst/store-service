package application

import (
	"context"
	"net/url"
	"strings"
	"time"

	"github.com/JIeeiroSst/notifyhub-service/internal/domain/model"
	"github.com/JIeeiroSst/notifyhub-service/internal/domain/port"
	"github.com/google/uuid"
)

type dataSourceService struct {
	sources port.DataSourceRepository
}

func NewDataSourceService(sources port.DataSourceRepository) port.DataSourceUsecase {
	return &dataSourceService{sources: sources}
}

func validateDataSource(ds *model.DataSource) error {
	if ds.Name == "" {
		return invalid("data source name is required")
	}
	u, err := url.Parse(ds.URL)
	if ds.URL == "" || err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return invalid("url must be an absolute http(s) URL")
	}
	ds.Method = strings.ToUpper(ds.Method)
	if ds.Method == "" {
		ds.Method = "GET"
	}
	switch strings.ToLower(ds.AuthType) {
	case "", "none", "bearer", "basic", "apikey":
	default:
		return invalid("auth_type must be none|bearer|basic|apikey")
	}
	return nil
}

func (s *dataSourceService) Create(ctx context.Context, ds *model.DataSource) (*model.DataSource, error) {
	if err := validateDataSource(ds); err != nil {
		return nil, err
	}
	now := time.Now()
	ds.ID = uuid.NewString()
	ds.CreatedAt, ds.UpdatedAt = now, now
	if err := s.sources.Create(ctx, ds); err != nil {
		return nil, err
	}
	return ds, nil
}

func (s *dataSourceService) Get(ctx context.Context, id string) (*model.DataSource, error) {
	return s.sources.Get(ctx, id)
}

func (s *dataSourceService) List(ctx context.Context) ([]*model.DataSource, error) {
	return s.sources.ListActive(ctx)
}

func (s *dataSourceService) Update(ctx context.Context, id string, apply func(*model.DataSource) error) (*model.DataSource, error) {
	ds, err := s.sources.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	createdAt := ds.CreatedAt
	if err := apply(ds); err != nil {
		return nil, invalid("%v", err)
	}
	ds.ID, ds.CreatedAt, ds.UpdatedAt = id, createdAt, time.Now()
	if err := validateDataSource(ds); err != nil {
		return nil, err
	}
	if err := s.sources.Update(ctx, ds); err != nil {
		return nil, err
	}
	return ds, nil
}

func (s *dataSourceService) Delete(ctx context.Context, id string) error {
	ds, err := s.sources.Get(ctx, id)
	if err != nil {
		return err
	}
	ds.IsActive = false
	ds.UpdatedAt = time.Now()
	return s.sources.Update(ctx, ds)
}
