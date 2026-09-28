package scheduler

import (
	"github.com/JIeeiroSst/notifyhub-service/internal/domain/port"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(New),
	fx.Provide(func(s *Scheduler) port.JobScheduler { return s }),
)
