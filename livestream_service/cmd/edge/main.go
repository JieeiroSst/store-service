package main

import (
	"github.com/JIeeiroSst/livestream-service/internal/infrastructure"
	"go.uber.org/fx"
)

func main() {
	app := fx.New(infrastructure.EdgeModule)
	app.Run()
}
