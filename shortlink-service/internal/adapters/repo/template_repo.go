package repo

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/shortlink-service/internal/domain"
	"gorm.io/gorm"
)

type TemplateRepo struct{ db *gorm.DB }

func NewTemplateRepo(db *gorm.DB) *TemplateRepo { return &TemplateRepo{db: db} }

func templateToModel(t *domain.LinkTemplate) *LinkTemplateModel {
	return &LinkTemplateModel{
		ID: t.ID, UserID: t.UserID, Name: t.Name, Slug: t.Slug, Description: t.Description,
		Settings: linkTemplateSettingsToJSON(t.Settings), IsDefault: t.IsDefault,
	}
}

func modelToTemplate(m *LinkTemplateModel) *domain.LinkTemplate {
	return &domain.LinkTemplate{
		ID: m.ID, UserID: m.UserID, Name: m.Name, Slug: m.Slug, Description: m.Description,
		Settings: jsonToLinkTemplateSettings(m.Settings), IsDefault: m.IsDefault,
		CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
	}
}

func (r *TemplateRepo) Create(ctx context.Context, tpl *domain.LinkTemplate) error {
	m := templateToModel(tpl)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	*tpl = *modelToTemplate(m)
	return nil
}

func (r *TemplateRepo) scoped(ctx context.Context, userID *string) *gorm.DB {
	q := r.db.WithContext(ctx)
	if userID != nil {
		q = q.Where("user_id = ?", *userID)
	}
	return q
}

func (r *TemplateRepo) List(ctx context.Context, userID *string) ([]*domain.LinkTemplate, error) {
	var models []LinkTemplateModel
	err := r.scoped(ctx, userID).Order("is_default DESC, name ASC").Find(&models).Error
	if err != nil {
		return nil, err
	}
	out := make([]*domain.LinkTemplate, len(models))
	for i := range models {
		out[i] = modelToTemplate(&models[i])
	}
	return out, nil
}

func (r *TemplateRepo) GetByID(ctx context.Context, id string, userID *string) (*domain.LinkTemplate, error) {
	var m LinkTemplateModel
	err := r.scoped(ctx, userID).Where("id = ?", id).Take(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return modelToTemplate(&m), nil
}

func (r *TemplateRepo) Update(ctx context.Context, id string, userID *string, patch map[string]interface{}) (*domain.LinkTemplate, error) {
	if len(patch) == 0 {
		return nil, errors.New("no updates provided")
	}
	res := r.scoped(ctx, userID).Model(&LinkTemplateModel{}).Where("id = ?", id).Updates(patch)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return r.GetByID(ctx, id, nil)
}

func (r *TemplateRepo) Delete(ctx context.Context, id string, userID *string) error {
	res := r.scoped(ctx, userID).Delete(&LinkTemplateModel{}, "id = ?", id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *TemplateRepo) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&LinkTemplateModel{}).Where("slug = ?", slug).Count(&count).Error
	return count > 0, err
}

func (r *TemplateRepo) SlugByID(ctx context.Context, id string) (string, error) {
	var slug string
	err := r.db.WithContext(ctx).Model(&LinkTemplateModel{}).Select("slug").Where("id = ?", id).Take(&slug).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", nil
	}
	return slug, err
}

func (r *TemplateRepo) LinkCountByTemplate(ctx context.Context, templateID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&LinkModel{}).Where("template_id = ?", templateID).Count(&count).Error
	return count, err
}

func (r *TemplateRepo) UnsetDefaults(ctx context.Context, userID *string, exceptID string) error {
	q := r.db.WithContext(ctx).Model(&LinkTemplateModel{})
	if userID != nil {
		q = q.Where("user_id = ?", *userID)
	} else {
		q = q.Where("user_id IS NULL")
	}
	if exceptID != "" {
		q = q.Where("id != ?", exceptID)
	}
	return q.Update("is_default", false).Error
}

func (r *TemplateRepo) SetDefault(ctx context.Context, id string) (*domain.LinkTemplate, error) {
	if err := r.db.WithContext(ctx).Model(&LinkTemplateModel{}).Where("id = ?", id).
		Updates(map[string]interface{}{"is_default": true, "updated_at": gorm.Expr("NOW()")}).Error; err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id, nil)
}
