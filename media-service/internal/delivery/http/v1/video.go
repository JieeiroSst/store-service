package v1

import (
	"strconv"

	"github.com/JIeeiroSst/media-service/dto"
	"github.com/JIeeiroSst/utils/response"
	"github.com/gin-gonic/gin"
)

func (h *Handler) initVideo(api *gin.RouterGroup) {
	videoGr := api.Group("/videos")
	{
		videoGr.POST("/upload", h.UploadVideo)
		videoGr.POST("/", h.SaveVideo)
		videoGr.POST("/search", h.SearchVideo)
		videoGr.GET("/:video_id", h.FindVideoByID)
	}
}

func (h *Handler) UploadVideo(ctx *gin.Context) {
	video, err := ctx.FormFile("video")
	if err != nil {
		response.ResponseStatus(ctx, 400, response.MessageStatus{
			Message: err.Error(),
			Error:   true,
		})
	}

	streamURL, err := h.usecase.Videos.UploadVideo(ctx, video)
	if err != nil {
		response.ResponseStatus(ctx, 500, response.MessageStatus{
			Message: err.Error(),
			Error:   true,
		})
	}
	response.ResponseStatus(ctx, 200, response.MessageStatus{
		Data:  streamURL,
		Error: false,
	})
}

func (h *Handler) SaveVideo(ctx *gin.Context) {
	var req dto.UploadVideoRequest
	if err := ctx.ShouldBind(&req); err != nil {
		response.ResponseStatus(ctx, 400, response.MessageStatus{
			Message: err.Error(),
			Error:   true,
		})
	}

	if err := h.usecase.Videos.SaveVideo(ctx, req); err != nil {
		response.ResponseStatus(ctx, 500, response.MessageStatus{
			Message: err.Error(),
			Error:   true,
		})
	}
	response.ResponseStatus(ctx, 200, response.MessageStatus{
		Error: false,
	})
}

func (h *Handler) SearchVideo(ctx *gin.Context) {
	var req dto.SearchVideoRequest
	if err := ctx.ShouldBind(&req); err != nil {
		response.ResponseStatus(ctx, 400, response.MessageStatus{
			Message: err.Error(),
			Error:   true,
		})
	}

	videos, err := h.usecase.Videos.SearchVideo(ctx, req)
	if err != nil {
		response.ResponseStatus(ctx, 500, response.MessageStatus{
			Message: err.Error(),
			Error:   true,
		})
	}
	response.ResponseStatus(ctx, 200, response.MessageStatus{
		Data:  videos,
		Error: false,
	})
}

func (h *Handler) FindVideoByID(ctx *gin.Context) {
	videoID, err := strconv.Atoi(ctx.Param("video_id"))
	if err != nil {
		response.ResponseStatus(ctx, 400, response.MessageStatus{
			Message: err.Error(),
			Error:   true,
		})
	}

	video, err := h.usecase.Videos.FindVideoByIDES(ctx, videoID)
	if err != nil {
		response.ResponseStatus(ctx, 500, response.MessageStatus{
			Message: err.Error(),
			Error:   true,
		})
	}
	response.ResponseStatus(ctx, 200, response.MessageStatus{
		Data:  video,
		Error: false,
	})
}
