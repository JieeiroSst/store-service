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
	FindViewByUser(ctx context.Context, userID, videoID int) (*model.View, error)
	FindByVideoID(ctx context.Context, viewID int) ([]model.View, error)
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
	view, err := r.FindViewByUser(ctx, req.UserID, req.ViewID)
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

func (r *ViewRepository) FindViewByUser(ctx context.Context, userID, videoID int) (*model.View, error) {
	var view *model.View
	if err := r.db.Where("user_id = ? ", userID).Where("video_id = ?", videoID).Find(&view).Error; err != nil {
		logger.Error(ctx, "FindByID error %v", err)
		return nil, err
	}
	return view, nil
}

func (r *ViewRepository) FindByVideoID(ctx context.Context, videoID int) ([]model.View, error) {
	var view []model.View
	if err := r.db.Where("video_id = ?", videoID).Find(&view).Error; err != nil {
		logger.Error(ctx, "FindByID error %v", err)
		return nil, err
	}
	return view, nil
}
