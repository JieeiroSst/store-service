package http

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/JIeeiroSst/video-service/config"
	"github.com/JIeeiroSst/video-service/internal/domain/port"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	usecase port.VideoUsecase
	cfg     *config.Config
}

func NewHandler(usecase port.VideoUsecase, cfg *config.Config) *Handler {
	return &Handler{usecase: usecase, cfg: cfg}
}

func (h *Handler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	size, _ := strconv.Atoi(c.Query("page_size"))
	sortBy := port.SortNewest
	if c.Query("sort") == string(port.SortPopular) {
		sortBy = port.SortPopular
	}
	videos, total, err := h.usecase.List(c.Request.Context(), port.ListQuery{
		Query: c.Query("q"), Sort: sortBy, Page: page, PageSize: size,
	})
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": videos, "total": total})
}

func (h *Handler) Related(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	videos, err := h.usecase.Related(c.Request.Context(), c.Param("id"), limit)
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": videos})
}

func (h *Handler) View(c *gin.Context) {
	if err := h.usecase.RecordView(c.Request.Context(), c.Param("id"), c.ClientIP()); err != nil {
		fail(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) Thumbnail(c *gin.Context) {
	a, err := h.usecase.Thumbnail(c.Request.Context(), c.Param("id"))
	if err != nil {
		fail(c, err)
		return
	}
	c.Header("Cache-Control", "public, max-age=86400")
	c.Data(http.StatusOK, a.ContentType, a.Data)
}

func (h *Handler) HLS(c *gin.Context) {
	a, err := h.usecase.HLS(c.Request.Context(), c.Param("id"), strings.TrimPrefix(c.Param("path"), "/"))
	if err != nil {
		fail(c, err)
		return
	}
	c.Header("Cache-Control", "public, max-age=31536000, immutable")
	c.Data(http.StatusOK, a.ContentType, a.Data)
}

func (h *Handler) Get(c *gin.Context) {
	v, err := h.usecase.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, v)
}

func (h *Handler) Upload(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, h.cfg.Upload.MaxBytes)
	mr, err := c.Request.MultipartReader()
	if err != nil {
		fail(c, fmt.Errorf("%w: expected multipart/form-data", port.ErrInvalid))
		return
	}

	var title, description string
	for {
		part, err := mr.NextPart()
		if err != nil {
			fail(c, fmt.Errorf("%w: missing \"file\" part", port.ErrInvalid))
			return
		}
		switch part.FormName() {
		case "title":
			b := make([]byte, 256)
			n, _ := part.Read(b)
			title = string(b[:n])
		case "description":
			b, _ := io.ReadAll(io.LimitReader(part, 8192))
			description = string(b)
		case "file":
			ct := part.Header.Get("Content-Type")
			if mt, _, err := mime.ParseMediaType(ct); err == nil {
				ct = mt
			}
			if title == "" {
				title = strings.TrimSuffix(part.FileName(), filepath.Ext(part.FileName()))
			}
			v, err := h.usecase.Upload(c.Request.Context(), port.UploadInput{
				Title: title, Description: description, ContentType: ct, Body: part,
			})
			if err != nil {
				var tooBig *http.MaxBytesError
				if errors.As(err, &tooBig) {
					c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "file too large"})
					return
				}
				fail(c, err)
				return
			}
			c.JSON(http.StatusCreated, v)
			return
		}
	}
}

func (h *Handler) Stream(c *gin.Context) {
	s, err := h.usecase.Open(c.Request.Context(), c.Param("id"))
	if err != nil {
		fail(c, err)
		return
	}
	hdr := c.Writer.Header()
	hdr.Set("Content-Type", s.Video.ContentType)
	hdr.Set("ETag", s.Video.ETag())
	hdr.Set("Cache-Control", "public, max-age=31536000, immutable")
	http.ServeContent(c.Writer, c.Request, "", s.Video.CreatedAt, s.Content)
}

func (h *Handler) Delete(c *gin.Context) {
	if err := h.usecase.Delete(c.Request.Context(), c.Param("id")); err != nil {
		fail(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func fail(c *gin.Context, err error) {
	switch {
	case errors.Is(err, port.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "video not found"})
	case errors.Is(err, port.ErrInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	}
}
