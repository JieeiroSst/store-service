package repository

import (
	"context"

	"github.com/JIeeiroSst/shortlink-service/model"
	"gorm.io/gorm"
)

type LinkClicks interface {
	GetLinkClicksByLinkID(ctx context.Context, linkID string) (model.LinkClick, error)
	CreateLinkClick(ctx context.Context, linkClick *model.LinkClick) error
}

type LinkClicksRepo struct {
	db *gorm.DB
}

func NewLinkClicksRepo(db *gorm.DB) *LinkClicksRepo {
	return &LinkClicksRepo{
		db: db,
	}
}

func (r *LinkClicksRepo) GetLinkClicksByLinkID(ctx context.Context, linkID string) (model.LinkClick, error) {
	var linkClick model.LinkClick
	if err := r.db.Where("link_id = ?", linkID).First(&linkClick).Error; err != nil {
		return linkClick, err
	}
	return linkClick, nil
}

func (r *LinkClicksRepo) CreateLinkClick(ctx context.Context, linkClick *model.LinkClick) error {
	if err := r.db.Create(linkClick).Error; err != nil {
		return err
	}
	return nil
}
