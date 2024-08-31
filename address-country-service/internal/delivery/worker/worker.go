package worker

import (
	"time"

	"gofr.dev/pkg/gofr"
)

type Worker struct {
}

func (w *Worker) Run(app gofr.App) {
	app.AddCronJob("* */5 * * *", "", func(ctx *gofr.Context) {
		ctx.Logger.Infof("current time is %v", time.Now())
	})

	app.AddCronJob("*/10 * * * * *", "", func(ctx *gofr.Context) {
		ctx.Logger.Infof("current time is %v", time.Now())
	})
}
