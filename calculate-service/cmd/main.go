package main

import (
	"github.com/JIeeiroSst/calculate-service/internal/infrastructure"
	"go.uber.org/fx"
)

func main() {
	fx.New(infrastructure.Module).Run()
}
