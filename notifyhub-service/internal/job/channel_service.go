package job

import (
	"context"
	"fmt"
	"time"

	"github.com/JIeeiroSst/notifyhub-service/internal/model"
	"github.com/JIeeiroSst/notifyhub-service/internal/store/mysql"
	tmplEngine "github.com/JIeeiroSst/notifyhub-service/internal/template"
	"github.com/google/uuid"
)

type ChannelService struct {
	store *mysql.Store
}

func NewChannelService(store *mysql.Store) *ChannelService {
	return &ChannelService{store: store}
}

func (s *ChannelService) Create(ctx context.Context, ch *model.Channel) (*model.Channel, error) {
	if ch.Name == "" {
		return nil, fmt.Errorf("channel name is required")
	}
	if ch.Type == "" {
		return nil, fmt.Errorf("channel type is required")
	}
	ch.ID = uuid.NewString()
	ch.CreatedAt = time.Now()
	ch.UpdatedAt = time.Now()
	if err := s.store.CreateChannel(ctx, ch); err != nil {
		return nil, err
	}
	return ch, nil
}

type DataSourceService struct {
	store *mysql.Store
}

func NewDataSourceService(store *mysql.Store) *DataSourceService {
	return &DataSourceService{store: store}
}

func (s *DataSourceService) Create(ctx context.Context, ds *model.DataSource) (*model.DataSource, error) {
	if ds.Name == "" {
		return nil, fmt.Errorf("data source name is required")
	}
	if ds.URL == "" {
		return nil, fmt.Errorf("url is required")
	}
	if ds.Method == "" {
		ds.Method = "GET"
	}
	ds.ID = uuid.NewString()
	ds.CreatedAt = time.Now()
	ds.UpdatedAt = time.Now()
	if err := s.store.CreateDataSource(ctx, ds); err != nil {
		return nil, err
	}
	return ds, nil
}

type TemplateService struct {
	store *mysql.Store
	tmpl  *tmplEngine.Engine
}

func NewTemplateService(store *mysql.Store, tmpl *tmplEngine.Engine) *TemplateService {
	return &TemplateService{store: store, tmpl: tmpl}
}

func (s *TemplateService) Create(ctx context.Context, t *model.Template) (*model.Template, error) {
	if t.Name == "" {
		return nil, fmt.Errorf("template name is required")
	}
	if t.Body == "" {
		return nil, fmt.Errorf("template body is required")
	}
	t.ID = uuid.NewString()
	t.IsActive = true
	t.CreatedAt = time.Now()
	t.UpdatedAt = time.Now()
	if err := s.store.CreateTemplate(ctx, t); err != nil {
		return nil, err
	}
	_ = s.tmpl.Compile(t)
	return t, nil
}

func (s *TemplateService) Update(ctx context.Context, id string, patch *model.Template) (*model.Template, error) {
	existing, err := s.store.GetTemplate(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("template not found: %w", err)
	}
	patch.ID = existing.ID
	patch.CreatedAt = existing.CreatedAt
	patch.UpdatedAt = time.Now()
	if err := s.store.UpdateTemplate(ctx, patch); err != nil {
		return nil, err
	}
	s.tmpl.Evict(id)
	_ = s.tmpl.Compile(patch)
	return patch, nil
}
