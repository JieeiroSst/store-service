package middleware

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TraceID(v string) zap.Field {
	return zap.String("trace_id", v)
}

// TraceIDFromContext returns the trace id stored in a standard context (for use in service layer).
func TraceIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(ContextTraceIDKey).(string)
	return v
}

const (
	ContextTraceIDKey = "trace_id"
	HeaderXRequestID  = "X-Request-ID"
	HeaderXTraceID    = "X-Trace-ID"
)

// Middleware attaches a UUID trace id to the Gin context.
// Priority: X-Request-ID header → X-Trace-ID header (must be valid UUID) → newly generated UUID.
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := resolveTraceID(c)
		c.Set(ContextTraceIDKey, traceID)
		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), ContextTraceIDKey, traceID))
		c.Writer.Header().Set(HeaderXRequestID, traceID)
		c.Next()
	}
}

// resolveTraceID picks a UUID from X-Request-ID or X-Trace-ID headers sent by the caller.
// Falls back to a freshly generated UUID when neither header is present or contains a valid UUID.
func resolveTraceID(c *gin.Context) string {
	for _, h := range []string{HeaderXRequestID, HeaderXTraceID} {
		if v := c.GetHeader(h); v != "" {
			if _, err := uuid.Parse(v); err == nil {
				return v
			}
		}
	}
	return uuid.NewString()
}

// GinTraceIDFromContext returns the trace id previously set by Middleware.
func GinTraceIDFromContext(c *gin.Context) string {
	if v, ok := c.Get(ContextTraceIDKey); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// GinLogger logs each request. Run after Tracing() so the trace ID is already set.
func GinLogger(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		tid := ""
		if v, ok := c.Get(ContextTraceIDKey); ok {
			tid, _ = v.(string)
		}

		start := time.Now()
		path := c.Request.URL.Path
		if c.Request.URL.RawQuery != "" {
			path += "?" + c.Request.URL.RawQuery
		}

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		fields := []zapcore.Field{
			TraceID(tid),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.Int("status", status),
			zap.Float64("latency_ms", float64(latency.Microseconds())/1000.0),
			zap.String("ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
		}

		if rid := c.GetHeader(HeaderXRequestID); rid != "" {
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
