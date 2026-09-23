package infrastructure

import (
	httpadapter "github.com/JIeeiroSst/payment-service/internal/adapter/primary/http"
	"github.com/JIeeiroSst/payment-service/internal/adapter/secondary/gateway"
	"github.com/JIeeiroSst/payment-service/internal/adapter/secondary/gateway/payoneer"
	"github.com/JIeeiroSst/payment-service/internal/adapter/secondary/gateway/paypal"
	"github.com/JIeeiroSst/payment-service/internal/adapter/secondary/gateway/stripe"
	"github.com/JIeeiroSst/payment-service/internal/adapter/secondary/gateway/wise"
	"github.com/JIeeiroSst/payment-service/internal/adapter/secondary/repository"
	"github.com/JIeeiroSst/payment-service/internal/application"
	"github.com/JIeeiroSst/payment-service/internal/infrastructure/database"
	"github.com/JIeeiroSst/payment-service/internal/infrastructure/server"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Invoke(initLogger),
	fx.Provide(newConfig),

	database.Module, // *gorm.DB

	repository.Module, // port.PaymentRepository

	paypal.Module,   // port.PaymentGateway (paypal), into group "payment_gateways"
	payoneer.Module, // port.PaymentGateway (payoneer), into group "payment_gateways"
	stripe.Module,   // port.PaymentGateway (stripe), into group "payment_gateways"
	wise.Module,     // port.PaymentGateway (wise), into group "payment_gateways"
	gateway.Module,  // port.PaymentGatewayResolver, aggregating the group above

	application.Module, // port.PaymentUsecase

	httpadapter.Module, // *httpadapter.Handler

	fx.Invoke(server.New),
)
