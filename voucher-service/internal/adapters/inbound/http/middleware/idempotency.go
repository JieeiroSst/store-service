package middleware

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"time"

	"github.com/JIeeiroSst/voucher-service/internal/platform/idempotency"
	"github.com/gin-gonic/gin"
)

const idempotencyKeyTTL = 24 * time.Hour

type bufferingWriter struct {
	gin.ResponseWriter
	buf *bytes.Buffer
}

func (w *bufferingWriter) Write(b []byte) (int, error) {
	w.buf.Write(b)
	return w.ResponseWriter.Write(b)
}

func Idempotency(store idempotency.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("Idempotency-Key")
		if key == "" {
			c.Next()
			return
		}

		bodyBytes, _ := io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		hash := sha256.Sum256(bodyBytes)
		requestHash := hex.EncodeToString(hash[:])

		claimed, err := store.Claim(c.Request.Context(), key, requestHash, idempotencyKeyTTL)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "internal_error", "message": "idempotency check failed"}})
			return
		}
		if !claimed {
			record, err := store.Get(c.Request.Context(), key)
			if err != nil || record == nil {
				c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": gin.H{"code": "duplicate_request", "message": "duplicate request"}})
				return
			}
			switch record.Status {
			case idempotency.StatusCompleted:
				c.Data(record.ResponseStatus, "application/json", record.ResponseBody)
				c.Abort()
			default:
				c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": gin.H{"code": "duplicate_request", "message": "a request with this idempotency key is already in progress"}})
			}
			return
		}

		buf := &bytes.Buffer{}
		writer := &bufferingWriter{ResponseWriter: c.Writer, buf: buf}
		c.Writer = writer

		c.Next()

		status := c.Writer.Status()
		if status >= 500 {
			_ = store.Release(c.Request.Context(), key)
			return
		}
		_ = store.Complete(c.Request.Context(), key, status, buf.Bytes())
	}
}
