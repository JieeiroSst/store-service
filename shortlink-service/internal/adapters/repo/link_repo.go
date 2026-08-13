package repo

import (
	"context"
	"errors"
	"time"

	"github.com/JIeeiroSst/shortlink-service/internal/domain"
	"github.com/JIeeiroSst/shortlink-service/internal/ports"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

var ErrNotFound = errors.New("not found")

type LinkRepo struct{ db *gorm.DB }

func NewLinkRepo(db *gorm.DB) *LinkRepo { return &LinkRepo{db: db} }

func (r *LinkRepo) Create(ctx context.Context, link *domain.Link) error {
	m := linkToModel(link)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	*link = *modelToLink(m)
	return nil
}

func (r *LinkRepo) scoped(ctx context.Context, filter ports.LinkFilter) *gorm.DB {
	q := r.db.WithContext(ctx)
	if filter.UserID != nil {
		q = q.Where("user_id = ?", *filter.UserID)
	}
	return q
}

func (r *LinkRepo) GetByID(ctx context.Context, id string, filter ports.LinkFilter) (*domain.Link, error) {
	var m LinkModel
	q := r.scoped(ctx, filter).Where("id = ?", id)
	if err := q.First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	link := modelToLink(&m)
	link.ClickCount, _ = r.clickCount(ctx, m.ID)
	link.TemplateSlug = r.templateSlugPtr(ctx, m.TemplateID)
	return link, nil
}

func (r *LinkRepo) clickCount(ctx context.Context, linkID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&ClickEventModel{}).Where("link_id = ?", linkID).Count(&count).Error
	return count, err
}

func (r *LinkRepo) templateSlugPtr(ctx context.Context, templateID *string) *string {
	if templateID == nil {
		return nil
	}
	slug, err := r.TemplateSlugByID(ctx, *templateID)
	if err != nil || slug == "" {
		return nil
	}
	return &slug
}

func (r *LinkRepo) List(ctx context.Context, filter ports.LinkFilter) ([]*domain.Link, error) {
	var models []LinkModel
	q := r.scoped(ctx, filter).Order("created_at DESC")
	if err := q.Find(&models).Error; err != nil {
		return nil, err
	}
	links := make([]*domain.Link, 0, len(models))
	for i := range models {
		l := modelToLink(&models[i])
		l.ClickCount, _ = r.clickCount(ctx, models[i].ID)
		l.TemplateSlug = r.templateSlugPtr(ctx, models[i].TemplateID)
		links = append(links, l)
	}
	return links, nil
}

// Update applies a partial patch (already snake_case DB column -> value)
// and returns the updated row. Mirrors the dynamic UPDATE builder in
// links.ts's PUT /api/links/:id.
func (r *LinkRepo) Update(ctx context.Context, id string, filter ports.LinkFilter, patch map[string]interface{}) (*domain.Link, error) {
	if len(patch) == 0 {
		return nil, errors.New("no updates provided")
	}
	q := r.scoped(ctx, filter).Model(&LinkModel{}).Where("id = ?", id)
	res := q.Updates(patch)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return r.GetByID(ctx, id, ports.LinkFilter{})
}

func (r *LinkRepo) Delete(ctx context.Context, id string, filter ports.LinkFilter) (*domain.Link, error) {
	var m LinkModel
	q := r.scoped(ctx, filter).Where("id = ?", id)
	if err := q.First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if err := r.db.WithContext(ctx).Delete(&LinkModel{}, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return modelToLink(&m), nil
}

func (r *LinkRepo) ExistsByShortCode(ctx context.Context, shortCode string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&LinkModel{}).Where("short_code = ?", shortCode).Count(&count).Error
	return count > 0, err
}

func (r *LinkRepo) TemplateSlugByID(ctx context.Context, templateID string) (string, error) {
	if templateID == "" {
		return "", nil
	}
	var slug string
	err := r.db.WithContext(ctx).Model(&LinkTemplateModel{}).
		Select("slug").Where("id = ?", templateID).Take(&slug).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", nil
	}
	return slug, err
}

type resolveRow struct {
	LinkModel
	TemplateSettings datatypes.JSON `gorm:"column:template_settings"`
	OrgSettings      datatypes.JSON `gorm:"column:org_settings"`
	OwnerSuspendedAt *time.Time     `gorm:"column:owner_suspended_at"`
}

func (r *LinkRepo) ResolveForRedirect(ctx context.Context, shortCode, templateSlug string) (*domain.Link, error) {
	q := r.db.WithContext(ctx).Table("links l").
		Select(`l.*, t.settings AS template_settings, o.settings AS org_settings, o.suspended_at AS owner_suspended_at`).
		Joins("LEFT JOIN link_templates t ON l.template_id = t.id").
		Joins("LEFT JOIN organizations o ON l.organization_id = o.id").
		Where("l.short_code = ?", shortCode).
		Where("l.is_active = true").
		Where("(l.expires_at IS NULL OR l.expires_at > NOW())")

	if templateSlug != "" {
		q = q.Where("t.slug = ?", templateSlug)
	}

	var row resolveRow
	if err := q.Take(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	link := modelToLink(&row.LinkModel)
	if len(row.TemplateSettings) > 0 {
		settings := jsonToLinkTemplateSettings(row.TemplateSettings)
		link.TemplateSettings = &settings
	}
	if len(row.OrgSettings) > 0 {
		settings := jsonToOrgSettings(row.OrgSettings)
		link.OrgSettings = &settings
	}
	if row.OwnerSuspendedAt != nil {
		link.OwnerSuspendedAt = row.OwnerSuspendedAt
	}
	if templateSlug != "" {
		link.TemplateSlug = &templateSlug
	}
	return link, nil
}
