package main

import (
	"github.com/JIeeiroSst/kitchen-service/internal/infrastructure"
	"go.uber.org/fx"
)

func main() {
	app := fx.New(infrastructure.Module)
	app.Run()
}
