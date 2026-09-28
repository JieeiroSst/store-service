package infrastructure

import (
	httpadapter "github.com/JIeeiroSst/shopify-service/internal/adapter/primary/http"
	"github.com/JIeeiroSst/shopify-service/internal/adapter/secondary/repository"
	"github.com/JIeeiroSst/shopify-service/internal/adapter/secondary/shopify"
	"github.com/JIeeiroSst/shopify-service/internal/application"
	"github.com/JIeeiroSst/shopify-service/internal/infrastructure/database"
	"github.com/JIeeiroSst/shopify-service/internal/infrastructure/server"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Invoke(initLogger),
	fx.Provide(newConfig),

	database.Module, // *gorm.DB

	repository.Module, // port.ProductRepository, port.OrderRepository, port.WebhookEventRepository

	shopify.Module, // port.ShopifyClient, port.WebhookVerifier

	application.Module, // port.ProductUsecase, port.OrderUsecase, port.WebhookUsecase

	httpadapter.Module, // *httpadapter.Handler

	fx.Invoke(server.New),
)
