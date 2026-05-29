package main

import (
	"go.uber.org/fx"

	"github.com/referral/service/internal/adapters/primary/http"
	dynamo "github.com/referral/service/internal/adapters/secondary/dynamodb"
	"github.com/referral/service/internal/config"
	"github.com/referral/service/internal/core/services"
	"github.com/referral/service/pkg/logger"
)

func main() {
	app := fx.New(
		// ── Infrastructure ───────────────────────────
		config.Module,
		logger.Module,

		// ── Secondary Adapters (DynamoDB repos) ──────
		dynamo.Module,

		// ── Core Business Logic ───────────────────────
		services.Module,

		// ── Primary Adapters (HTTP) ────────────────────
		http.Module,
		http.ServerModule,
	)

	app.Run()
}
