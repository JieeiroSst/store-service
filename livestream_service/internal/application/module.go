package application

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewRoomUsecase),
	fx.Provide(NewNodeSchedulerUsecase),
	fx.Provide(NewIngestUsecase),
	fx.Provide(NewPublishUsecase),
	fx.Provide(NewViewerUsecase),
	fx.Provide(NewChatUsecase),
	fx.Provide(NewModerationUsecase),
)
