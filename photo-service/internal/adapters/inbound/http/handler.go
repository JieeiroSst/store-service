package http

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/JIeeiroSst/photo-service/internal/application/ports"
	"github.com/JIeeiroSst/photo-service/internal/domain"
)

const (
	maxUploadMemory = 32 << 20 // 32 MiB kept in memory before spilling to temp files
	maxRawBodySize  = 32 << 20 // 32 MiB cap on a raw (non-multipart) image buffer
)

type CompositionHandler struct {
	useCase ports.ComposeImageUseCase
}

func NewCompositionHandler(useCase ports.ComposeImageUseCase) *CompositionHandler {
	return &CompositionHandler{useCase: useCase}
}

func (h *CompositionHandler) Create(c *gin.Context) {
	if isMultipart(c) {
		h.createFromMultipart(c)
		return
	}
	h.createFromBuffer(c)
}

func (h *CompositionHandler) createFromMultipart(c *gin.Context) {
	if err := c.Request.ParseMultipartForm(maxUploadMemory); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid multipart form: " + err.Error()})
		return
	}

	var dto layoutDTO
	if raw := c.PostForm("layout"); raw != "" {
		if err := json.Unmarshal([]byte(raw), &dto); err != nil {
			c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid layout json: " + err.Error()})
			return
		}
	}

	sources, err := h.collectSources(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	h.compose(c, sources, dto)
}

func (h *CompositionHandler) createFromBuffer(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxRawBodySize)

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(c.Request.Body); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "failed to read request body: " + err.Error()})
		return
	}
	if buf.Len() == 0 {
		c.JSON(http.StatusBadRequest, errorResponse{Error: domain.ErrNoSources.Error()})
		return
	}

	var dto layoutDTO
	if raw := c.GetHeader("X-Layout"); raw != "" {
		if err := json.Unmarshal([]byte(raw), &dto); err != nil {
			c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid layout json: " + err.Error()})
			return
		}
	}

	sources := []domain.ImageSource{{
		Type:        domain.ImageSourceTypeUpload,
		Data:        buf.Bytes(),
		ContentType: c.ContentType(),
		Order:       0,
	}}

	h.compose(c, sources, dto)
}

func (h *CompositionHandler) compose(c *gin.Context, sources []domain.ImageSource, dto layoutDTO) {
	cmd := ports.ComposeImagesCommand{
		Sources:        sources,
		Layout:         dto.toDomain(),
		IdempotencyKey: c.GetHeader("Idempotency-Key"),
	}

	job, err := h.useCase.ComposeImages(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(statusForError(err), errorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, newCompositionResponse(job))
}

func isMultipart(c *gin.Context) bool {
	return strings.HasPrefix(c.ContentType(), "multipart/form-data")
}

// Get handles GET /v1/compositions/{id}.
func (h *CompositionHandler) Get(c *gin.Context) {
	id := c.Param("id")
	job, err := h.useCase.GetComposition(c.Request.Context(), id)
	if err != nil {
		c.JSON(statusForError(err), errorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, newCompositionResponse(job))
}

func (h *CompositionHandler) collectSources(c *gin.Context) ([]domain.ImageSource, error) {
	var sources []domain.ImageSource
	order := 0

	if c.Request.MultipartForm != nil {
		for _, fh := range c.Request.MultipartForm.File["images"] {
			f, err := fh.Open()
			if err != nil {
				return nil, err
			}
			data, readErr := io.ReadAll(f)
			f.Close()
			if readErr != nil {
				return nil, readErr
			}
			contentType := fh.Header.Get("Content-Type")
			sources = append(sources, domain.ImageSource{
				Type:        domain.ImageSourceTypeUpload,
				Data:        data,
				ContentType: contentType,
				Order:       order,
			})
			order++
		}
	}

	for _, u := range c.PostFormArray("urls") {
		if u == "" {
			continue
		}
		sources = append(sources, domain.ImageSource{
			Type:  domain.ImageSourceTypeURL,
			URL:   u,
			Order: order,
		})
		order++
	}

	for _, key := range c.PostFormArray("object_keys") {
		if key == "" {
			continue
		}
		sources = append(sources, domain.ImageSource{
			Type:      domain.ImageSourceTypeMinIO,
			ObjectKey: key,
			Order:     order,
		})
		order++
	}

	if len(sources) == 0 {
		return nil, domain.ErrNoSources
	}
	return sources, nil
}
