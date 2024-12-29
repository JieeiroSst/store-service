package build

import (
	"github.com/JIeeiroSst/media-service/dto"
	"github.com/JIeeiroSst/media-service/model"
	"github.com/JIeeiroSst/utils/geared_id"
	"github.com/JIeeiroSst/utils/time_custom"
)

func BuildSaveView(view dto.View) model.View {
	if view.ViewID == 0 {
		view.ViewID = geared_id.GearedIntID()
	}

	return model.View{
		ViewID:    view.ViewID,
		UserID:    view.UserID,
		VideoID:   view.VideoID,
		Platform:  view.Platform,
		CreatedAt: time_custom.TimeInCountry("Ho_Chi_Minh"),
	}
}

func BuildView(view *model.View) *dto.View {
	if view == nil {
		return nil
	}
	return &dto.View{
		ViewID:    view.ViewID,
		UserID:    view.UserID,
		VideoID:   view.VideoID,
		Platform:  view.Platform,
		CreatedAt: view.CreatedAt,
		TotalView: view.TotalView,
	}
}

func BuildViews(view []model.View) []dto.View {
	views := make([]dto.View, 0)

	for _, v := range view {
		views = append(views, *BuildView(&v))
	}
	return views
}
