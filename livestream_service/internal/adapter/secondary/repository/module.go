package repository

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewRoomRepository),
	fx.Provide(NewStreamRepository),
	fx.Provide(NewVODRepository),
)
