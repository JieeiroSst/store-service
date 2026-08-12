package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/JIeeiroSst/cdn-service/internal/domain"
	"github.com/JIeeiroSst/cdn-service/internal/ports"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	usecase ports.FileUsecase
}

func NewHandler(usecase ports.FileUsecase) *Handler {
	return &Handler{usecase: usecase}
}

func (h *Handler) Presign(c *gin.Context) {
	var req presignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pu, err := h.usecase.CreatePresignedUpload(c.Request.Context(), ports.CreatePresignInput{
		FileName:    req.FileName,
		ContentType: req.ContentType,
		SizeBytes:   req.SizeBytes,
		UploadedBy:  c.GetHeader("X-Uploaded-By"),
	})
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusCreated, toPresignResponse(pu))
}

func (h *Handler) Confirm(c *gin.Context) {
	id, err := parseFileID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	file, err := h.usecase.ConfirmUpload(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, toFileResponse(*file))
}

func (h *Handler) Get(c *gin.Context) {
	id, err := parseFileID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	file, err := h.usecase.GetFile(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, toFileResponse(*file))
}

func (h *Handler) Download(c *gin.Context) {
	id, err := parseFileID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	direct := c.Query("direct") == "true"

	url, err := h.usecase.GetDownloadURL(c.Request.Context(), id, direct)
	if err != nil {
		writeError(c, err)
		return
	}
	c.Redirect(http.StatusFound, url)
}

func (h *Handler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))

	files, err := h.usecase.ListFiles(c.Request.Context(), ports.ListFilesInput{Limit: limit, Offset: offset})
	if err != nil {
		writeError(c, err)
		return
	}

	resp := make([]fileResponse, len(files))
	for i, f := range files {
		resp[i] = toFileResponse(f)
	}
	c.JSON(http.StatusOK, listFilesResponse{Files: resp})
}

func (h *Handler) Delete(c *gin.Context) {
	id, err := parseFileID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.usecase.DeleteFile(c.Request.Context(), id); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func parseFileID(c *gin.Context) (uuid.UUID, error) {
	return uuid.Parse(c.Param("id"))
}

func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrFileNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrInvalidStatus), errors.Is(err, domain.ErrInvalidFileInput):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrFileTooLarge):
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
