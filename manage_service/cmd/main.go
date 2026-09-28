package main

import (
	"github.com/JIeeiroSst/manage-service/internal/infrastructure"
	"go.uber.org/fx"
)

func main() {
	fx.New(
		infrastructure.Module,
		fx.WithLogger(infrastructure.FxLogger),
		fx.StopTimeout(infrastructure.StopTimeout),
	).Run()
}
