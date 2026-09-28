package http

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

type RouterConfig struct {
	APIKey      string
	RateLimiter *RateLimiter
	Log         *zap.Logger
}

func NewRouter(h *Handler, cfg RouterConfig) *gin.Engine {
	r := gin.New()
	r.Use(
		Recovery(cfg.Log),
		RequestID(),
		Logger(cfg.Log),
		CORS(),
		MaxBodySize(4<<20), // 4 MB
	)

	r.GET("/health", h.Health)
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	v1 := r.Group("/api/v1", cfg.RateLimiter.Middleware(), APIKeyAuth(cfg.APIKey))
	{
		ch := v1.Group("/channels")
		{
			ch.POST("", h.CreateChannel)
			ch.GET("", h.ListChannels)
			ch.GET("/:id", h.GetChannel)
			ch.PUT("/:id", h.UpdateChannel)
			ch.DELETE("/:id", h.DeleteChannel)
		}

		ds := v1.Group("/data-sources")
		{
			ds.POST("", h.CreateDataSource)
			ds.GET("", h.ListDataSources)
			ds.GET("/:id", h.GetDataSource)
			ds.PUT("/:id", h.UpdateDataSource)
			ds.DELETE("/:id", h.DeleteDataSource)
		}

		tmpl := v1.Group("/templates")
		{
			tmpl.POST("", h.CreateTemplate)
			tmpl.GET("", h.ListTemplates)
			tmpl.GET("/:id", h.GetTemplate)
			tmpl.PUT("/:id", h.UpdateTemplate)
			tmpl.DELETE("/:id", h.DeleteTemplate)
		}

		jobs := v1.Group("/jobs")
		{
			jobs.POST("", h.CreateJob)
			jobs.GET("", h.ListJobs)
			jobs.GET("/:id", h.GetJob)
			jobs.PUT("/:id", h.UpdateJob)
			jobs.DELETE("/:id", h.DeleteJob)

			jobs.POST("/:id/pause", h.PauseJob)
			jobs.POST("/:id/resume", h.ResumeJob)
			jobs.POST("/:id/trigger", h.TriggerJob)
		}

		v1.GET("/history", h.ListHistory)
		v1.GET("/scheduler/status", h.SchedulerStatus)
	}

	return r
}
