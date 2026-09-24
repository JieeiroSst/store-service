package http

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/JIeeiroSst/customer-relationship-service/common"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/model"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/port"
	"github.com/gin-gonic/gin"
)

const (
	defaultLimit = 50
	maxLimit     = 200
)

type Resource interface {
	Register(api *gin.RouterGroup)
}

type crudHandler[T any, PT model.Entity[T]] struct {
	path    string
	uc      port.Usecase[T]
	filters []string
}

func (h *crudHandler[T, PT]) Register(api *gin.RouterGroup) {
	g := api.Group(h.path)
	g.POST("", h.create)
	g.GET("", h.list)
	g.GET("/:id", h.get)
	g.PUT("/:id", h.update)
	g.DELETE("/:id", h.delete)
}

func (h *crudHandler[T, PT]) create(c *gin.Context) {
	var entity T
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	PT(&entity).SetID(0)

	result, err := h.uc.Create(c.Request.Context(), &entity)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (h *crudHandler[T, PT]) get(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	result, err := h.uc.Get(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *crudHandler[T, PT]) list(c *gin.Context) {
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", strconv.Itoa(defaultLimit)))
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 || limit > maxLimit {
		limit = defaultLimit
	}

	q := port.ListQuery{Offset: offset, Limit: limit, Search: strings.TrimSpace(c.Query("q"))}
	for _, name := range h.filters {
		if v := c.Query(name); v != "" {
			if q.Equals == nil {
				q.Equals = map[string]any{}
			}
			q.Equals[name] = filterValue(v)
		}
	}

	result, total, err := h.uc.List(c.Request.Context(), q)
	if err != nil {
		writeError(c, err)
		return
	}
	c.Header("X-Total-Count", strconv.FormatInt(total, 10))
	c.JSON(http.StatusOK, result)
}

func filterValue(v string) any {
	if n, err := strconv.ParseUint(v, 10, 64); err == nil {
		return n
	}
	return v
}

func (h *crudHandler[T, PT]) update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var entity T
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := h.uc.Update(c.Request.Context(), id, &entity)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *crudHandler[T, PT]) delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.uc.Delete(c.Request.Context(), id); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func parseID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return 0, false
	}
	return uint(id), true
}

func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, common.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, common.ErrInvalidRequest):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, common.ErrMalicious):
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
	case errors.Is(err, common.ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, common.ErrTooLarge):
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": err.Error()})
	case errors.Is(err, common.ErrUnsupportedMedia):
		c.JSON(http.StatusUnsupportedMediaType, gin.H{"error": err.Error()})
	case errors.Is(err, common.ErrUpstream):
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
	case errors.Is(err, common.ErrInvalidSignature):
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
	case errors.Is(err, common.ErrInvalidTransition):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
