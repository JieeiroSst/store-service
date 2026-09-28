package repository

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewTxManager),
	fx.Provide(NewMedicineRepository),
	fx.Provide(NewBatchRepository),
	fx.Provide(NewMovementRepository),
	fx.Provide(NewInteractionRepository),
	fx.Provide(NewDispenseRepository),
	fx.Provide(NewTerminologyRepository),
)
