package redisstore

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewNodeRegistry),
	fx.Provide(NewViewerCounter),
	fx.Provide(NewChatBroadcaster),
	fx.Provide(NewModerationStore),
)
