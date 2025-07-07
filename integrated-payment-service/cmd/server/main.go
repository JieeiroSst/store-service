package main

import (
	"log"

	"github.com/JIeeiroSst/integrated-payment-service/internal/application/presentation/handlers"
	"github.com/JIeeiroSst/integrated-payment-service/internal/application/presentation/routes"
	"github.com/JIeeiroSst/integrated-payment-service/internal/application/services"
	"github.com/JIeeiroSst/integrated-payment-service/internal/config"
	"github.com/JIeeiroSst/integrated-payment-service/internal/domain/entities"
	"github.com/JIeeiroSst/integrated-payment-service/internal/domain/interfaces"
	"github.com/JIeeiroSst/integrated-payment-service/internal/infrastructure/database"
	"github.com/JIeeiroSst/integrated-payment-service/internal/infrastructure/payments"
	"github.com/JIeeiroSst/integrated-payment-service/pkg/logger"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	logger := logger.New(cfg.LogLevel)

	db, err := database.Connect(cfg.Database.URL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	db.AutoMigrate(&entities.Payment{}, &entities.User{}, &entities.Transaction{})

	paymentRepo := database.NewPaymentRepository(db)

	processors := make(map[entities.PaymentMethod]interfaces.PaymentProcessor)
	processors[entities.PaymentMethodMomo] = payments.NewMoMoProcessor(payments.MoMoConfig(cfg.MoMo))
	processors[entities.PaymentMethodVNPay] = payments.NewVNPayProcessor(payments.VNPayConfig(cfg.VNPay))

	paymentService := services.NewPaymentService(paymentRepo, processors, logger)

	paymentHandler := handlers.NewPaymentHandler(paymentService)

	r := gin.Default()
	routes.SetupRoutes(r, paymentHandler)

	log.Printf("Server starting on port %s", cfg.Server.Port)
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
