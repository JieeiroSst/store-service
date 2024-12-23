package repository

import (
	"context"

	"github.com/JIeeiroSst/media-service/model"
	"github.com/JIeeiroSst/utils/logger"
	"gorm.io/gorm"
)

type View interface {
	SaveView(ctx context.Context, view model.View) error
	FindByID(ctx context.Context, viewID int) (*model.View, error)
}

type ViewRepository struct {
	db *gorm.DB
}

func NewViewRepository(db *gorm.DB) *ViewRepository {
	return &ViewRepository{
		db: db,
	}
}

func (r *ViewRepository) SaveView(ctx context.Context, req model.View) error {
	view, err := r.FindByID(ctx, req.ViewID)
	if err != nil {
		logger.Error(ctx, "FindView error %v", err)
		return err
	}
	if view == nil || view.ViewID == 0 {
		if err := r.db.Create(&req).Error; err != nil {
			logger.Error(ctx, "CreateView error %v", err)
			return err
		}
		return nil
	}

	view.TotalView += 1
	if err := r.db.Save(&view).Error; err != nil {
		logger.Error(ctx, "SaveView error %v", err)
		return err
	}
	return nil
}

func (r *ViewRepository) FindByID(ctx context.Context, viewID int) (*model.View, error) {
	var view *model.View
	if err := r.db.Where("view_id = ?", viewID).Find(&view).Error; err != nil {
		logger.Error(ctx, "FindByID error %v", err)
		return nil, err
	}
	return view, nil
}
