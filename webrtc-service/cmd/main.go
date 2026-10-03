package main

import (
	"github.com/JIeeiroSst/webrtc-service/internal/infrastructure"
	"go.uber.org/fx"
)

func main() {
	app := fx.New(infrastructure.Module)
	app.Run()
}
