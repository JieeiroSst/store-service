package main

import (
	"go.uber.org/fx"

	"github.com/JIeerioSst/rent-house-service/internal/bootstrap"
)

func main() {
	fx.New(bootstrap.Module).Run()
}
