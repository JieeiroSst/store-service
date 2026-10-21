package main

import (
	"go.uber.org/fx"

	"github.com/JIeeiroSst/wallet-service/di"
)

func main() {
	fx.New(di.App).Run()
}
