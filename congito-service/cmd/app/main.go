package main

import (
	"log/slog"

	"github.com/JIeeiroSst/congito-service/internal/api"
	"github.com/JIeeiroSst/congito-service/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	slog.Info("env parsed successfully", "environment", cfg.Env)

	api.Run(cfg)
}
