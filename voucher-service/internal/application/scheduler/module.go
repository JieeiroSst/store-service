package scheduler

import (
	"context"

	"go.uber.org/fx"
)

func registerJobsOnStart(svc SchedulerService) {
	if err := svc.RegisterJobs(context.Background()); err != nil {
		panic(err)
	}
}

var Module = fx.Module("scheduler-app",
	fx.Provide(fx.Annotate(NewService, fx.As(new(SchedulerService)))),
	fx.Invoke(registerJobsOnStart),
)
