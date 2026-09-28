package main

import (
	"github.com/JIeeiroSst/notifyhub-service/internal/infrastructure"
	"go.uber.org/fx"
)

func main() {
	app := fx.New(
		infrastructure.Module,
		fx.WithLogger(infrastructure.FxLogger),
		fx.StopTimeout(infrastructure.StopTimeout),
	)
	app.Run()
}
