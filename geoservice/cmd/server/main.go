package main

import (
	"go.uber.org/fx"

	"github.com/geoservice/internal/app"
)

func main() {
	fx.New(app.ServerModule).Run()
}
