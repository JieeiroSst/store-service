package infrastructure

import (
	httpadapter "github.com/JIeeiroSst/doordash-service/internal/adapter/primary/http"
	"github.com/JIeeiroSst/doordash-service/internal/adapter/secondary/notifierclient"
	"github.com/JIeeiroSst/doordash-service/internal/adapter/secondary/paymentclient"
	"github.com/JIeeiroSst/doordash-service/internal/adapter/secondary/repository"
	"github.com/JIeeiroSst/doordash-service/internal/adapter/secondary/restaurantclient"
	"github.com/JIeeiroSst/doordash-service/internal/adapter/secondary/userclient"
	"github.com/JIeeiroSst/doordash-service/internal/application"
	"github.com/JIeeiroSst/doordash-service/internal/infrastructure/database"
	"github.com/JIeeiroSst/doordash-service/internal/infrastructure/server"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Invoke(initLogger),
	fx.Provide(newConfig),

	database.Module, // *gorm.DB

	repository.Module,       // port.OrderRepository, port.OrderTrackingRepository, port.DriverAssignmentRepository
	userclient.Module,       // port.UserClient        -> user_service
	restaurantclient.Module, // port.RestaurantClient  -> restaurant/menu catalog service
	paymentclient.Module,    // port.PaymentClient     -> payment service
	notifierclient.Module,   // port.NotifierClient    -> notification service

	application.Module, // port.OrderUsecase, port.TrackingUsecase, port.DriverAssignmentUsecase

	httpadapter.Module, // *httpadapter.Handler

	fx.Invoke(server.New),
)
