package application

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewMedicineService),
	fx.Provide(NewInventoryService),
	fx.Provide(NewInteractionService),
	fx.Provide(NewDispenseService),
	fx.Provide(NewTerminologyService),
)
