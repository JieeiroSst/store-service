package postgres

import (
	"github.com/JIeeiroSst/vending-machine-service/internal/port"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(NewDB,
			fx.As(new(port.Transactor)),
			fx.As(new(port.HealthChecker)),
			fx.As(fx.Self()),
		),
		NewMachineRepository,
		NewMaintenanceRepository,
		NewEventRepository,
		NewCategoryRepository,
		NewProductRepository,
		NewInventoryRepository,
		NewSessionRepository,
		NewReservationRepository,
		NewPaymentRepository,
		NewOrderRepository,
		NewReportRepository,
	),
)
