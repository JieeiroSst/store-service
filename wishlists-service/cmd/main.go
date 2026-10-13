package main

import (
	"github.com/JIeeiroSst/wishlists-service/internal/infrastructure"
	"go.uber.org/fx"
)

func main() {
	fx.New(infrastructure.Options()...).Run()
}
