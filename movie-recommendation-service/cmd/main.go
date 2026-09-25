package main

import (
	"github.com/JIeeiroSst/movie-recommendation-service/internal/infrastructure"
	"go.uber.org/fx"
)

func main() {
	fx.New(infrastructure.Module).Run()
}
