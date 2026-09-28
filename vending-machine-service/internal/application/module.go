package application

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(
		NewMachineService,
		NewCatalogService,
		NewInventoryService,
		NewVendingService,
		NewAlertService,
	),
)
