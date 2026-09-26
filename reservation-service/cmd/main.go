package main

import (
	"go.uber.org/fx"

	"github.com/JIeeiroSSt/reservation-service/internal/bootstrap"
)

func main() {
	fx.New(bootstrap.Module).Run()
}
