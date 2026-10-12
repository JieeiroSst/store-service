package main

import (
	"github.com/JIeeiroSst/ejabberd-service/internal/infrastructure"
	"go.uber.org/fx"
)

func main() {
	app := fx.New(infrastructure.Module, infrastructure.Logger())
	app.Run()
}
