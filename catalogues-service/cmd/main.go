package main

import (
	"github.com/JIeeiroSst/catalogues-service/internal/infrastructure"
	"go.uber.org/fx"
)

func main() {
	fx.New(infrastructure.Module).Run()
}
