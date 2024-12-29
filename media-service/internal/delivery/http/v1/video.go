package v1

import (
	"github.com/JIeeiroSst/utils/response"
	"github.com/gin-gonic/gin"
)

func (h *Handler) initVideo(api *gin.RouterGroup) {


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
	response.ResponseStatus(ctx, 500, response.MessageStatus{
		Data:  streamURL,
		Error: false,
	})
}
