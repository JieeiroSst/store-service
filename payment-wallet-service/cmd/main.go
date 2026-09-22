package main

import (
	"github.com/JIeeiroSst/payment-wallet-service/internal/infrastructure"
	"go.uber.org/fx"
)

func main() {
	app := fx.New(infrastructure.Module)
	app.Run()
}
