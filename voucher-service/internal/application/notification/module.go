package notification

import "go.uber.org/fx"

var Module = fx.Module("notification-app",
	fx.Provide(fx.Annotate(NewService, fx.As(new(NotificationService)))),
)
