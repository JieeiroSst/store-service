package repository

import (
	"context"

	"github.com/JIeeiroSst/shortlink-service/model"
	"gorm.io/gorm"
)

type Links interface {
	CreateLink(ctx context.Context, link *model.Link) error
	GetLinkByShortCode(ctx context.Context, shortCode string) (*model.Link, error)
	GetLinkByID(ctx context.Context, id string) (*model.Link, error)
	DeleteLink(ctx context.Context, shortCode string) error
	CreateOrUpdateLinkClick(ctx context.Context, linkID string) error
	GetLinks(ctx context.Context, page, pageSize int) ([]model.Link, int64, error)
}

type LinkRepo struct {
	repo *gorm.DB
}

func NewLinkRepo(repo *gorm.DB) *LinkRepo {
	return &LinkRepo{
		repo: repo,
	}
}

func (r *LinkRepo) CreateLink(ctx context.Context, link *model.Link) error {
	if err := r.repo.Create(link).Error; err != nil {
		return err
	}
	return nil
}

func (r *LinkRepo) GetLinkByShortCode(ctx context.Context, shortCode string) (*model.Link, error) {
	var link model.Link
	if err := r.repo.Where("short_code = ?", shortCode).First(&link).Error; err != nil {
		return nil, err
	}
	return &link, nil
}

func (r *LinkRepo) GetLinkByID(ctx context.Context, id string) (*model.Link, error) {
	var link model.Link
	if err := r.repo.Where("id = ?", id).First(&link).Error; err != nil {
		return nil, err
	}
	return &link, nil
}

func (r *LinkRepo) DeleteLink(ctx context.Context, shortCode string) error {
	if err := r.repo.Where("short_code = ?", shortCode).Delete(&model.Link{}).Error; err != nil {
		return err
	}
	return nil
}

func (r *LinkRepo) CreateOrUpdateLinkClick(ctx context.Context, linkID string) error {
	var linkClick model.Link
	if err := r.repo.Where("id = ?", linkID).First(&linkClick).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			linkClick.ID = linkID
			linkClick.TotalClicks = 1
			if err := r.repo.Create(&linkClick).Error; err != nil {
				return err
			}
		}
		return err
	}

	linkClick.TotalClicks++
	linkClick.ID = linkID
	if err := r.repo.Save(&linkClick).Error; err != nil {
		return err
	}
	return nil
}

func (r *LinkRepo) GetLinks(ctx context.Context, page, pageSize int) ([]model.Link, int64, error) {
	var links []model.Link
	var total int64
	if page == 0 {
		page = 1
	}

	if pageSize == 0 {
		pageSize = 10
	}

	if err := r.repo.Model(&model.Link{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := r.repo.Offset((page - 1) * pageSize).Limit(pageSize).Find(&links).Error; err != nil {
		return nil, 0, err
	}
	return links, total, nil
}
