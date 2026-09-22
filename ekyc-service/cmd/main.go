package main

import (
	"go.uber.org/fx"

	"github.com/JIeeiroSst/ekyc-service/internal/infrastructure"
)

func main() {
	app := fx.New(infrastructure.Module)
	app.Run()
}
