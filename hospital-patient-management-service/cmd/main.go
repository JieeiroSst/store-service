package main

import (
	"github.com/JIeeroSst/hospital-patient-management-service/internal/infrastructure"
	"go.uber.org/fx"
)

func main() {
	fx.New(infrastructure.Module).Run()
}
