package httpapi

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/JIeeiroSst/wallet-service/internal/domain/model"
)

const idempotencyHeader = "Idempotency-Key"

type errorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

type page[T any] struct {
	Items    []T   `json:"items"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}

func statusFor(kind model.ErrorKind) int {
	switch kind {
	case model.KindInvalid:
		return http.StatusBadRequest
	case model.KindNotFound:
		return http.StatusNotFound
	case model.KindConflict:
		return http.StatusConflict
	case model.KindUnprocessable:
		return http.StatusUnprocessableEntity
	case model.KindForbidden:
		return http.StatusForbidden
	case model.KindLocked:
		return http.StatusLocked
	case model.KindUpstream:
		return http.StatusBadGateway
	}
	return http.StatusInternalServerError
}

func (a *API) fail(c *gin.Context, err error) {
	e := model.AsError(err)
	status := statusFor(e.Kind)
	if status == http.StatusInternalServerError {
		a.logger.Error("request failed", zap.String("path", c.FullPath()), zap.Error(err))
	}
	c.AbortWithStatusJSON(status, errorResponse{Error: e.Message, Code: e.Code})
}

func badRequest(c *gin.Context, message string) {
	c.AbortWithStatusJSON(http.StatusBadRequest, errorResponse{Error: message, Code: "INVALID_ARGUMENT"})
}

func bind(c *gin.Context, dst any) bool {
	if err := c.ShouldBindJSON(dst); err != nil {
		badRequest(c, err.Error())
		return false
	}
	return true
}

type ids struct {
	err error
}

func (p *ids) parse(name, raw string, required bool) uuid.UUID {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		if required && p.err == nil {
			p.err = fmt.Errorf("%s is required", name)
		}
		return uuid.Nil
	}
	id, err := uuid.Parse(raw)
	if err != nil && p.err == nil {
		p.err = fmt.Errorf("%s must be a valid UUID", name)
	}
	return id
}

func (p *ids) required(name, raw string) uuid.UUID { return p.parse(name, raw, true) }

func (p *ids) optional(name, raw string) uuid.UUID { return p.parse(name, raw, false) }

func (p *ids) ok(c *gin.Context) bool {
	if p.err != nil {
		badRequest(c, p.err.Error())
		return false
	}
	return true
}

func pathID(c *gin.Context, name string) (uuid.UUID, bool) {
	var p ids
	id := p.required(name, c.Param(name))
	return id, p.ok(c)
}

func idempotencyKey(c *gin.Context, fallback string) string {
	if key := strings.TrimSpace(c.GetHeader(idempotencyHeader)); key != "" {
		return key
	}
	return strings.TrimSpace(fallback)
}

func pagination(c *gin.Context) (int, int) {
	p, _ := strconv.Atoi(c.Query("page"))
	size, _ := strconv.Atoi(c.Query("page_size"))
	return model.NormalizePage(p, size)
}

func parseTime(raw string) (*time.Time, error) {
	if raw == "" {
		return nil, nil
	}
	for _, layout := range []string{time.RFC3339Nano, "2006-01-02"} {
		if t, err := time.Parse(layout, raw); err == nil {
			t = t.UTC()
			return &t, nil
		}
	}
	return nil, fmt.Errorf("invalid time %q, use RFC3339 or YYYY-MM-DD", raw)
}

func timeRange(c *gin.Context) (*time.Time, *time.Time, bool) {
	from, err := parseTime(c.Query("from"))
	if err != nil {
		badRequest(c, err.Error())
		return nil, nil, false
	}
	to, err := parseTime(c.Query("to"))
	if err != nil {
		badRequest(c, err.Error())
		return nil, nil, false
	}
	return from, to, true
}

func csv(raw string) []string {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.ToUpper(strings.TrimSpace(p)); p != "" {
			out = append(out, p)
		}
	}
	return out
}
