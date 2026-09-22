package main

import (
	"github.com/JIeeiroSst/partner-service/internal/app"
	"go.uber.org/fx"
)

func main() {
	fx.New(app.Module).Run()
}
