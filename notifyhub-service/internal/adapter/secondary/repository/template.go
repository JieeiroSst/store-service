package repository

import (
	"context"

	"github.com/JIeeiroSst/notifyhub-service/internal/domain/model"
	"github.com/JIeeiroSst/notifyhub-service/internal/domain/port"
	"gorm.io/gorm"
)

type templateRepository struct {
	db *gorm.DB
}

func NewTemplateRepository(db *gorm.DB) port.TemplateRepository {
	return &templateRepository{db: db}
}

func (r *templateRepository) Create(ctx context.Context, t *model.Template) error {
	return translate(r.db.WithContext(ctx).Create(t).Error, "template "+t.Name)
}

func (r *templateRepository) Get(ctx context.Context, id string) (*model.Template, error) {
	var t model.Template
	if err := r.db.WithContext(ctx).First(&t, "id = ?", id).Error; err != nil {
		return nil, translate(err, "template "+id)
	}
	return &t, nil
}

func (r *templateRepository) ListActive(ctx context.Context, channel string) ([]*model.Template, error) {
	q := r.db.WithContext(ctx).Where("is_active = ?", true)
	if channel != "" {
		q = q.Where("channel = ?", channel)
	}
	var ts []*model.Template
	return ts, q.Order("created_at desc").Find(&ts).Error
}

func (r *templateRepository) Update(ctx context.Context, t *model.Template) error {
	return affected(r.db.WithContext(ctx).Model(t).Select("*").Omit("created_at").Updates(t), "template "+t.ID)
}
