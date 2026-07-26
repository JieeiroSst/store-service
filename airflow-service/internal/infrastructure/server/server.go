package server

import (
	"context"

	httpadapter "github.com/JIeeiroSst/airflow-service/internal/adapter/primary/http"
	"go.uber.org/fx"
	"gofr.dev/pkg/gofr"
)

type Params struct {
	fx.In

	LC      fx.Lifecycle
	Handler *httpadapter.Handler
}

func New(p Params) {
	app := gofr.New()

	httpadapter.RegisterRoutes(app, p.Handler)

	p.LC.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go app.Run()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return app.Shutdown(ctx)
		},
	})
}
