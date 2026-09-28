package http

import (
	"errors"
	"log"
	"net/http"
	"path/filepath"
	"slices"
	"strconv"
	"time"

	"github.com/JIeeiroSst/draw-image-service/config"
	"github.com/JIeeiroSst/draw-image-service/internal/domain"
	"github.com/JIeeiroSst/draw-image-service/internal/port"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	collage        port.CollageService
	maxUploadBytes int64
}

func NewHandler(collage port.CollageService, cfg *config.Config) *Handler {
	return &Handler{collage: collage, maxUploadBytes: cfg.Server.MaxUploadBytes}
}

type collageResponse struct {
	ID          string    `json:"id"`
	URL         string    `json:"url"`
	Bucket      string    `json:"bucket"`
	ObjectKey   string    `json:"object_key"`
	ContentType string    `json:"content_type"`
	Size        int64     `json:"size"`
	Width       int       `json:"width"`
	Height      int       `json:"height"`
	ImageCount  int       `json:"image_count"`
	Layout      string    `json:"layout"`
	Columns     int       `json:"columns"`
	CreatedAt   time.Time `json:"created_at"`
}

func toResponse(c *domain.Collage) collageResponse {
	return collageResponse{
		ID:          c.ID,
		URL:         "/images/" + c.ID,
		Bucket:      c.Bucket,
		ObjectKey:   c.ObjectKey,
		ContentType: c.ContentType,
		Size:        c.Size,
		Width:       c.Width,
		Height:      c.Height,
		ImageCount:  c.ImageCount,
		Layout:      string(c.Layout),
		Columns:     c.Columns,
		CreatedAt:   c.CreatedAt,
	}
}

func (h *Handler) Upload(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, h.maxUploadBytes)

	layout, err := domain.ParseLayout(c.Query("layout"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	columns := 0
	if raw := c.Query("columns"); raw != "" {
		if columns, err = strconv.Atoi(raw); err != nil || columns <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "columns must be a positive integer"})
			return
		}
	}

	form, err := c.MultipartForm()
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "request body too large"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid multipart form: " + err.Error()})
		return
	}
	defer form.RemoveAll()

	centers := form.File["center"]
	if len(centers) > 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at most one center file"})
		return
	}
	files := slices.Concat(centers, form.File["files"])
	sources := make([]domain.Source, 0, len(files))
	for _, fh := range files {
		f, err := fh.Open()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "open " + fh.Filename + ": " + err.Error()})
			return
		}
		defer f.Close()
		sources = append(sources, domain.Source{Name: filepath.Base(fh.Filename), Content: f})
	}

	result, err := h.collage.Create(c.Request.Context(), domain.CollageRequest{
		Sources: sources,
		Layout:  layout,
		Columns: columns,
	})
	if err != nil {
		writeError(c, err)
		return
	}

	c.Header("Location", "/images/"+result.ID)
	c.JSON(http.StatusCreated, toResponse(result))
}

func (h *Handler) GetImage(c *gin.Context) {
	result, rc, err := h.collage.Open(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeError(c, err)
		return
	}
	defer rc.Close()

	c.DataFromReader(http.StatusOK, result.Size, result.ContentType, rc, map[string]string{
		"Cache-Control": "public, max-age=31536000, immutable",
		"ETag":          `"` + result.ID + `"`,
	})
}

func (h *Handler) GetInfo(c *gin.Context) {
	result, err := h.collage.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, toResponse(result))
}

func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func writeError(c *gin.Context, err error) {
	status := statusFor(err)
	if status == http.StatusInternalServerError {
		log.Printf("%s %s: %v", c.Request.Method, c.Request.URL.Path, err)
		c.JSON(status, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(status, gin.H{"error": err.Error()})
}

func statusFor(err error) int {
	switch {
	case errors.Is(err, domain.ErrNoImages),
		errors.Is(err, domain.ErrInvalidLayout),
		errors.Is(err, domain.ErrInvalidImage),
		errors.Is(err, domain.ErrInvalidID):
		return http.StatusBadRequest
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrTooManyImages),
		errors.Is(err, domain.ErrImageTooLarge):
		return http.StatusRequestEntityTooLarge
	}
	return http.StatusInternalServerError
}
