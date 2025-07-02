package middleware

import (
	"bytes"
	"io"
	"time"

	"github.com/JIeeiroSst/kms/models"
	"github.com/JIeeiroSst/kms/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (rw responseWriter) Write(b []byte) (int, error) {
	rw.body.Write(b)
	return rw.ResponseWriter.Write(b)
}

func AuditMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		rw := &responseWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBufferString(""),
		}
		c.Writer = rw

		c.Next()

		userID, _ := c.Get("user_id")
		username, _ := c.Get("username")

		var actorID uuid.UUID
		var actorName string

		if userID != nil {
			actorID = userID.(uuid.UUID)
		}
		if username != nil {
			actorName = username.(string)
		}

		level := models.LogLevelInfo
		success := c.Writer.Status() < 400
		if !success {
			level = models.LogLevelError
		}

		metadata := map[string]interface{}{
			"method":        c.Request.Method,
			"path":          c.Request.URL.Path,
			"status_code":   c.Writer.Status(),
			"duration_ms":   time.Since(start).Milliseconds(),
			"request_size":  len(requestBody),
			"response_size": rw.body.Len(),
		}

		auditLog := models.AuditLog{
			ID:         uuid.New(),
			ActorID:    actorID,
			ActorName:  actorName,
			Action:     c.Request.Method + " " + c.Request.URL.Path,
			Resource:   "api",
			ResourceID: c.Request.URL.Path,
			Timestamp:  start,
			IPAddress:  c.ClientIP(),
			UserAgent:  c.Request.UserAgent(),
			Success:    success,
			Metadata:   metadata,
			Level:      level,
		}

		if !success && rw.body.Len() > 0 {
			auditLog.ErrorMsg = rw.body.String()
		}

		go func() {
			services.SaveAuditLog(auditLog)
		}()
	}
}
