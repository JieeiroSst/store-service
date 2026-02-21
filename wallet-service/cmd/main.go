package main

import (
	"go.uber.org/fx"

	"github.com/JIeeiroSst/wallet-service/di"
)

func main() {
	app := fx.New(di.App)
	app.Run()
}
