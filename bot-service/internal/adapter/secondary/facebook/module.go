package facebook

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewClient),
	fx.Provide(fx.Annotate(NewMessageSender, fx.ResultTags(`group:"channel_senders"`))),
	fx.Provide(fx.Annotate(NewContentPublisher, fx.ResultTags(`group:"content_publishers"`))),
)
