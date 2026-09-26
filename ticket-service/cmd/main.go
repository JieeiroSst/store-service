package main

import (
	"go.uber.org/fx"

	"github.com/JIeeiroSst/ticket-service/internal/bootstrap"
)

func main() {
	fx.New(bootstrap.Module).Run()
}
