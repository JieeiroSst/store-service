package api

import (
	"github.com/JIeeiroSst/notifyhub-service/internal/api/handler"
	"github.com/JIeeiroSst/notifyhub-service/internal/api/middleware"
	"github.com/JIeeiroSst/notifyhub-service/internal/scheduler"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func NewRouter(
	h *handler.Handler,
	sched *scheduler.Scheduler,
	apiKey string,
	rateLimit int,
	log *zap.Logger,
) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	rl := middleware.NewRateLimiter(rateLimit)

	r.Use(
		middleware.Recovery(log),
		handler.RequestID(),
		middleware.Logger(log),
		middleware.CORS(),
		middleware.MaxBodySize(4<<20), // 4 MB
		rl.Middleware(),
	)

	r.GET("/health", h.Health)
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	auth := middleware.APIKeyAuth(apiKey)
	v1 := r.Group("/api/v1", auth)
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

		v1.GET("/scheduler/status", handler.SchedulerStatus(sched))
	}

	return r
}
