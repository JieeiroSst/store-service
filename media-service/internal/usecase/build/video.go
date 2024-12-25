package build

import (
	"github.com/JIeeiroSst/media-service/dto"
	"github.com/JIeeiroSst/media-service/model"
)

func BuildVideo(video model.Video) dto.Video {
	return dto.Video{
		VideoID:      video.VideoID,
		Description:  video.Description,
		ThumbnailURL: video.ThumbnailURL,
		StreamURL:    video.StreamURL,
		UploadedOn:   video.UploadedOn,
	}
}

func BuildVideos(req []model.Video) []dto.Video {
	videos := make([]dto.Video, 0)

	for _, v := range req {
		videos = append(videos, BuildVideo(v))
	}

	return videos
}

func BuildSearchVideo(req model.SearchVideo) dto.SearchVideo {
	return dto.SearchVideo{
		Videos: BuildVideos(req.Videos),
		Total:  req.Total,
		Page:   req.Page,
		Size:   req.Size,
		Pages:  req.Pages,
	}
}
