package notifier

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewPushSender),
	fx.Provide(NewEmailSender),
	fx.Provide(NewSlackSender),
)
