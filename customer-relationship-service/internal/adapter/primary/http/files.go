package http

import (
	"errors"
	"io"
	"mime"
	"net/http"
	"strconv"
	"strings"

	"github.com/JIeeiroSst/customer-relationship-service/config"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/model"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/port"
	"github.com/gin-gonic/gin"
)

const multipartOverhead = 1 << 20

type fileHandler struct {
	files   port.ContractFileUsecase
	maxBody int64
}

func NewFileHandler(files port.ContractFileUsecase, cfg *config.Config) *fileHandler {
	max := cfg.Files.MaxBytes
	if max <= 0 {
		max = 25 << 20
	}
	return &fileHandler{files: files, maxBody: max + multipartOverhead}
}

func (h *fileHandler) Register(api *gin.RouterGroup) {
	api.GET("/contract-file-kinds", func(c *gin.Context) { c.JSON(http.StatusOK, model.FileKinds) })

	g := api.Group("/contracts/:id/files")
	g.POST("", h.upload)
	g.GET("", h.list)
	g.GET("/:file_id", h.get)
	g.GET("/:file_id/download", h.download)
	g.GET("/:file_id/history", h.history)
	g.DELETE("/:file_id", h.delete)

	api.GET("/contracts/:id/file-events", h.audit)
}

func fileID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("file_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file id"})
		return 0, false
	}
	return uint(id), true
}

func (h *fileHandler) upload(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, h.maxBody)

	header, err := c.FormFile("file")
	if err != nil {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "file too large"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "a multipart form with a \"file\" part is required"})
		return
	}
	f, err := header.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var replaces uint
	if v := strings.TrimSpace(c.PostForm("replaces")); v != "" {
		n, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "replaces must be a file id"})
			return
		}
		replaces = uint(n)
	}

	result, err := h.files.Upload(c.Request.Context(), port.FileUpload{
		ContractID:  id,
		Kind:        strings.TrimSpace(c.PostForm("kind")),
		Name:        header.Filename,
		Description: c.PostForm("description"),
		Data:        data,
		ReplacesID:  replaces,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (h *fileHandler) list(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", strconv.Itoa(defaultLimit)))
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 || limit > maxLimit {
		limit = defaultLimit
	}
	rows, total, err := h.files.List(c.Request.Context(), id, port.FileListFilter{
		Kind: c.Query("kind"), AllVersions: c.Query("all_versions") == "true", Offset: offset, Limit: limit,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.Header("X-Total-Count", strconv.FormatInt(total, 10))
	c.JSON(http.StatusOK, rows)
}

func (h *fileHandler) get(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	fid, ok := fileID(c)
	if !ok {
		return
	}
	f, err := h.files.Get(c.Request.Context(), id, fid)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, f)
}

func (h *fileHandler) download(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	fid, ok := fileID(c)
	if !ok {
		return
	}
	f, data, err := h.files.Download(c.Request.Context(), id, fid)
	if err != nil {
		writeError(c, err)
		return
	}
	if disposition := mime.FormatMediaType("attachment", map[string]string{"filename": f.Name}); disposition != "" {
		c.Header("Content-Disposition", disposition)
	}
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("Cache-Control", "private, no-store")
	c.Data(http.StatusOK, f.ContentType, data)
}

func (h *fileHandler) delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	fid, ok := fileID(c)
	if !ok {
		return
	}
	if err := h.files.Delete(c.Request.Context(), id, fid); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *fileHandler) history(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	fid, ok := fileID(c)
	if !ok {
		return
	}
	result, err := h.files.History(c.Request.Context(), id, fid)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *fileHandler) audit(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", strconv.Itoa(defaultLimit)))
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 || limit > maxLimit {
		limit = defaultLimit
	}
	f := port.FileAuditFilter{Action: c.Query("action"), Offset: offset, Limit: limit}
	if v := c.Query("file_id"); v != "" {
		n, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "file_id must be a file id"})
			return
		}
		f.FileID = uint(n)
	}
	rows, total, err := h.files.Audit(c.Request.Context(), id, f)
	if err != nil {
		writeError(c, err)
		return
	}
	c.Header("X-Total-Count", strconv.FormatInt(total, 10))
	c.JSON(http.StatusOK, rows)
}
