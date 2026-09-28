package notifier

import (
	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/port"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(NewPushSender),
	fx.Provide(func(s *pushSender) port.PushSender { return s }),
	fx.Provide(func(s *pushSender) port.MulticastSender { return s }),
	fx.Provide(func(s *pushSender) port.TokenValidator { return s }),
	fx.Provide(NewEmailSender),
	fx.Provide(func(s *emailSender) port.EmailSender { return s }),
	fx.Provide(func(s *emailSender) port.BatchEmailSender { return s }),
	fx.Provide(NewSlackSender),
)
