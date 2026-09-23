package main

import (
	"os"

	"github.com/JIeeiroSst/video-service/internal/infrastructure"
	"go.uber.org/fx"
)

func main() {
	module := infrastructure.APIModule
	if os.Getenv("ROLE") == "worker" {
		module = infrastructure.WorkerModule
	}
	fx.New(module).Run()
}
