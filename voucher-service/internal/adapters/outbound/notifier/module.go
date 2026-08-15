package notifier

import (
	notificationapp "github.com/JIeeiroSst/voucher-service/internal/application/notification"
	"go.uber.org/fx"
)

var Module = fx.Module("notifier",
	fx.Provide(
		fx.Annotate(NewSMTPNotifier, fx.As(new(notificationapp.Notifier)), fx.ResultTags(`group:"notifiers"`)),
		fx.Annotate(NewSMSNotifier, fx.As(new(notificationapp.Notifier)), fx.ResultTags(`group:"notifiers"`)),
		fx.Annotate(NewRegistry, fx.ParamTags(`group:"notifiers"`), fx.As(new(notificationapp.NotifierRegistry))),
	),
)
