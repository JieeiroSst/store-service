package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/JIeeiroSst/order-processing-service/internal/activity"
	"github.com/JIeeiroSst/order-processing-service/internal/config"
	"github.com/JIeeiroSst/order-processing-service/internal/domain/usecase"
	infraHTTP "github.com/JIeeiroSst/order-processing-service/internal/infrastructure/http"
	infraMySQL "github.com/JIeeiroSst/order-processing-service/internal/infrastructure/mysql"
	infraTemporal "github.com/JIeeiroSst/order-processing-service/internal/infrastructure/temporal"
	"github.com/JIeeiroSst/order-processing-service/internal/proxy"
	"github.com/JIeeiroSst/order-processing-service/internal/worker"
	"github.com/JIeeiroSst/order-processing-service/pkg/logger"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
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

	var paid *activity.PaidOrderActivities
	if cfg.OrderDB != "" {
		db, err := gorm.Open(mysql.Open(cfg.OrderDB), &gorm.Config{TranslateError: true})
		if err != nil {
			log.Fatalf("Failed to open order database: %v", err)
		}
		store, err := infraMySQL.NewPaidOrderStore(db)
		if err != nil {
			log.Fatalf("Failed to prepare order tables: %v", err)
		}
		parcel := activity.DefaultParcel()
		parcel.ItemWeightGram = cfg.Services.ParcelItemWeightGram
		paid = &activity.PaidOrderActivities{Store: store, Notifier: notifyProxy, Parcel: parcel}
		if cfg.Services.ShippingAPIKey != "" {
			paid.Shipments = proxy.NewShipmentProxy(cfg.Services.ShippingBaseURL, cfg.Services.ShippingAPIKey)
		} else {
			logger.Logger.Warn("SHIPPING_SERVICE_API_KEY is not set: paid orders are marked 'Needs shipping' instead of creating GHN shipments")
		}
	} else {
		logger.Logger.Warn("ORDER_DB_DSN is not set: PaidOrderWorkflow is not registered")
	}

	// Create and start worker
	w := worker.NewOrderWorker(temporalClient, activities, paid)

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
