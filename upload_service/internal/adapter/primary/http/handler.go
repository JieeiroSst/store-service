package http

import (
	"errors"
	"io"
	"mime"
	"net/http"
	"strconv"
	"strings"

	"github.com/JIeeiroSst/upload-service/config"
	"github.com/JIeeiroSst/upload-service/internal/domain/model"
	"github.com/JIeeiroSst/upload-service/internal/domain/port"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	files       port.FileUsecase
	maxBody     int64
	serviceOnly []string
}

func NewHandler(files port.FileUsecase, cfg *config.Config) *Handler {
	return &Handler{files: files, maxBody: cfg.Upload.MaxBytes + 1<<20, serviceOnly: cfg.Auth.ServiceOnlyPrefixes}
}

func (h *Handler) restricted(receiverID string) bool {
	for _, p := range h.serviceOnly {
		if strings.HasPrefix(receiverID, p) {
			return true
		}
	}
	return false
}

func (h *Handler) allowReceiver(c *gin.Context, receiverID string) bool {
	if len(h.serviceOnly) == 0 || isService(c) || !h.restricted(receiverID) {
		return true
	}
	c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
	return false
}

func (h *Handler) allowFile(c *gin.Context, id string) bool {
	if len(h.serviceOnly) == 0 || isService(c) {
		return true
	}
	f, err := h.files.Get(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return false
	}
	if h.restricted(f.ReceiverID) {
		writeError(c, model.ErrNotFound)
		return false
	}
	return true
}

const routePrefix = "/api/v1/upload"

func contentURL(id string) string { return routePrefix + "/" + id + "/content" }

func present(f *model.File) *model.File {
	out := *f
	out.URL = contentURL(f.ID)
	return &out
}

func (h *Handler) upload(c *gin.Context) (port.Upload, func(), error) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, h.maxBody)
	fh, err := c.FormFile("file")
	if err != nil {
		fh, err = c.FormFile("image")
	}
	if err != nil {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			return port.Upload{}, nil, err
		}
		return port.Upload{}, nil, model.Invalid("multipart form field 'file' is required")
	}
	f, err := fh.Open()
	if err != nil {
		return port.Upload{}, nil, model.Invalid("unreadable upload")
	}
	return port.Upload{ReceiverID: c.Query("receiver_id"), FileName: fh.Filename, Body: f}, func() { _ = f.Close() }, nil
}

func (h *Handler) Create(c *gin.Context) {
	in, closeBody, err := h.upload(c)
	if err != nil {
		writeError(c, err)
		return
	}
	defer closeBody()
	if !h.allowReceiver(c, in.ReceiverID) {
		return
	}
	f, err := h.files.Create(c.Request.Context(), in)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, present(f))
}

func (h *Handler) Replace(c *gin.Context) {
	if !h.allowFile(c, c.Param("id")) {
		return
	}
	in, closeBody, err := h.upload(c)
	if err != nil {
		writeError(c, err)
		return
	}
	defer closeBody()
	f, err := h.files.Replace(c.Request.Context(), c.Param("id"), in)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, present(f))
}

func (h *Handler) Get(c *gin.Context) {
	if !h.allowFile(c, c.Param("id")) {
		return
	}
	f, err := h.files.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, present(f))
}

func (h *Handler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))
	if !h.allowReceiver(c, c.Query("receiver_id")) {
		return
	}
	files, total, err := h.files.List(c.Request.Context(), c.Query("receiver_id"), limit, offset)
	if err != nil {
		writeError(c, err)
		return
	}
	items := make([]*model.File, len(files))
	for i := range files {
		items[i] = present(&files[i])
	}
	c.JSON(http.StatusOK, gin.H{"files": items, "total_count": total})
}

func (h *Handler) Content(c *gin.Context) {
	if !h.allowFile(c, c.Param("id")) {
		return
	}
	d, err := h.files.Open(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeError(c, err)
		return
	}
	defer d.Body.Close()

	c.Header("Content-Type", d.File.ContentType)
	c.Header("Content-Length", strconv.FormatInt(d.File.Size, 10))
	c.Header("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": d.File.FileName}))
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("Cache-Control", "private, no-store")
	c.Status(http.StatusOK)
	if _, err := io.Copy(c.Writer, d.Body); err != nil {
		logrus.WithError(err).WithField("file_id", d.File.ID).Warn("download interrupted")
	}
}

func (h *Handler) Delete(c *gin.Context) {
	if !h.allowFile(c, c.Param("id")) {
		return
	}
	if err := h.files.Delete(c.Request.Context(), c.Param("id")); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func writeError(c *gin.Context, err error) {
	var maxErr *http.MaxBytesError
	switch {
	case errors.As(err, &maxErr), errors.Is(err, model.ErrTooLarge):
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": err.Error()})
	case errors.Is(err, model.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
	case errors.Is(err, model.ErrUnsupportedType):
		c.JSON(http.StatusUnsupportedMediaType, gin.H{"error": err.Error()})
	case errors.Is(err, model.ErrInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, model.ErrUpstream):
		logrus.WithError(err).Error("dependency unavailable")
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "storage is unavailable"})
	default:
		logrus.WithError(err).Error("request failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	}
}
