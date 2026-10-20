package main

import (
	"github.com/JIeeiroSst/lucky-service/internal/infrastructure"
	"go.uber.org/fx"
)

func main() {
	fx.New(infrastructure.Options()...).Run()
}
