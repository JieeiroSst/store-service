package main

import (
	"github.com/JIeeiroSst/post-service/internal/infrastructure"
	"go.uber.org/fx"
)

func main() {
	app := fx.New(infrastructure.Module)
	app.Run()
}
