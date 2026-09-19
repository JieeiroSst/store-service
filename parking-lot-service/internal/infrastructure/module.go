package infrastructure

import (
	httpadapter "github.com/JIeeiroSst/parking-lot-service/internal/adapter/primary/http"
	"github.com/JIeeiroSst/parking-lot-service/internal/adapter/secondary/repository"
	"github.com/JIeeiroSst/parking-lot-service/internal/application"
	"github.com/JIeeiroSst/parking-lot-service/internal/infrastructure/database"
	"github.com/JIeeiroSst/parking-lot-service/internal/infrastructure/server"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Invoke(initLogger),
	fx.Provide(newConfig),

	database.Module, // *gorm.DB

	repository.Module, // port.VehicleRepository, port.ParkingSpotRepository, port.TicketRepository, port.RateRepository, port.PaymentRepository

	application.Module, // port.ParkingUsecase, port.RateUsecase

	httpadapter.Module, // *httpadapter.Handler

	fx.Invoke(server.New),
)
