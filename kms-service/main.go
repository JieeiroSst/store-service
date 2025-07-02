package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"github.com/JIeeiroSst/kms/config"
	"github.com/JIeeiroSst/kms/controllers"
	"github.com/JIeeiroSst/kms/middleware"
	"github.com/JIeeiroSst/kms/models"
	"github.com/JIeeiroSst/kms/services"
	"github.com/JIeeiroSst/kms/storage"
	"github.com/JIeeiroSst/kms/utils"
)

func main() {
	// Load configuration
	config.LoadConfig()

	// Initialize master key
	if err := utils.InitMasterKey(config.AppConfig.MasterKeyPath); err != nil {
		log.Fatal("Failed to initialize master key:", err)
	}

	// Initialize database
	db, err := storage.NewPostgresDB()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Initialize cache
	cache, err := storage.NewRedisCache()
	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}
	defer cache.Close()

	// Initialize services
	services.InitServices(db, cache)

	// Initialize scheduler
	scheduler := services.NewScheduler(db, cache)
	scheduler.Start()
	defer scheduler.Stop()

	// Setup Gin router
	if config.AppConfig.LogLevel == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// Global middleware
	r.Use(middleware.RateLimitMiddleware(rate.Limit(config.AppConfig.RateLimitPerMin), config.AppConfig.RateLimitPerMin))
	r.Use(middleware.AuditMiddleware())

	// Public routes
	r.GET("/health", controllers.HealthCheck)

	// Protected routes
	api := r.Group("/api/v1")
	api.Use(middleware.AuthMiddleware())

	// Key management routes
	keys := api.Group("/keys")
	{
		keys.POST("", middleware.RequirePermission("key:create"), controllers.CreateKeyV2)
		keys.GET("", middleware.RequirePermission("key:list"), controllers.ListKeys)
		keys.GET("/:id", middleware.RequirePermission("key:read"), controllers.GetKey)
		keys.GET("/:id/use", middleware.RequirePermission("key:use"), controllers.GetKeyForUse)
		keys.GET("/:id/stats", middleware.RequirePermission("key:read"), controllers.GetKeyUsageStats)
		keys.PUT("/:id/rotate", middleware.RequirePermission("key:rotate"), controllers.RotateKeyV2)
		keys.DELETE("/:id", middleware.RequirePermission("key:delete"), controllers.DeleteKey)
		keys.GET("/:id/audit", middleware.RequirePermission("audit:read"), controllers.GetKeyAuditLogs)
	}

	// Audit routes
	audit := api.Group("/audit")
	audit.Use(middleware.RequireRole(models.RoleAdmin, models.RoleAuditor))
	{
		audit.GET("/logs", controllers.GetAuditLogsV2)
	}

	// Start server
	server := &http.Server{
		Addr:         ":" + config.AppConfig.ServerPort,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("KMS Service starting on port %s", config.AppConfig.ServerPort)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("Failed to start server:", err)
	}
}
