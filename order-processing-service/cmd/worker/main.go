package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/JIeeiroSst/order-processing-service/internal/activity"
	"github.com/JIeeiroSst/order-processing-service/internal/config"
	infraHTTP "github.com/JIeeiroSst/order-processing-service/internal/infrastructure/http"
	infraTemporal "github.com/JIeeiroSst/order-processing-service/internal/infrastructure/temporal"
	"github.com/JIeeiroSst/order-processing-service/internal/domain/usecase"
	"github.com/JIeeiroSst/order-processing-service/internal/proxy"
	"github.com/JIeeiroSst/order-processing-service/internal/worker"
	"github.com/JIeeiroSst/order-processing-service/pkg/logger"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize logger
	logger.Init(cfg.Env)
	defer logger.Sync()

	// Create Temporal client
	temporalClient, err := infraTemporal.NewClient(cfg.Temporal)
	if err != nil {
		log.Fatalf("Failed to create Temporal client: %v", err)
	}
	defer temporalClient.Close()

	// Initialize proxy layer (internal API clients)
	paymentProxy := proxy.NewPaymentProxy(cfg.Services.PaymentBaseURL, cfg.Services.PaymentAPIKey)
	inventoryProxy := proxy.NewInventoryProxy(cfg.Services.InventoryBaseURL, cfg.Services.InventoryAPIKey)
	shippingProxy := proxy.NewShippingProxy(cfg.Services.ShippingBaseURL, cfg.Services.ShippingAPIKey)
	notifyProxy := proxy.NewNotificationProxy(cfg.Services.NotificationBaseURL, cfg.Services.NotificationAPIKey)

	// Initialize repository (in-memory for demo; swap with real DB impl)
	orderRepo := infraHTTP.NewInMemoryOrderRepository()

	// Initialize use case
	orderUseCase := usecase.NewOrderUseCase(orderRepo, paymentProxy, inventoryProxy, shippingProxy, notifyProxy)

	// Initialize activities
	activities := activity.NewOrderActivities(orderUseCase)

	// Create and start worker
	w := worker.NewOrderWorker(temporalClient, activities)

	logger.Logger.Infow("Starting Temporal Worker",
		"taskQueue", "ORDER_PROCESSING_TASK_QUEUE",
		"namespace", cfg.Temporal.Namespace,
	)

	if err := w.Start(); err != nil {
		log.Fatalf("Failed to start worker: %v", err)
	}
	defer w.Stop()

	logger.Logger.Info("Worker started successfully. Listening for tasks...")

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Logger.Info("Shutting down worker...")
}
