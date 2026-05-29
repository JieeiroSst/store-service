// middleware.go — Gin HTTP request/response logger using the zap core.
// Logs method, path, status, latency, client IP, and request-id on every request.
package logger

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func GinMiddleware(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		if c.Request.URL.RawQuery != "" {
			path += "?" + c.Request.URL.RawQuery
		}

		c.Next()
		latency := time.Since(start)
		status := c.Writer.Status()

		fields := []zapcore.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.Int("status", status),
			zap.Float64("latency_ms", float64(latency.Microseconds())/1000.0),
			zap.String("ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
		}

		if rid := c.GetHeader("X-Request-ID"); rid != "" {
			fields = append(fields, zap.String("request_id", rid))
		}

		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("error", c.Errors.ByType(gin.ErrorTypePrivate).String()))
		}

		switch {
		case status >= 500:
			log.Error("request", fields...)
		case status >= 400:
			log.Warn("request", fields...)
		default:
			log.Info("request", fields...)
		}
	}
}
