package httpadapter

import (
	"net/http"

	"go.uber.org/fx"

	"github.com/JIeeiroSst/toggle-service/internal/adapter/primary/http/handler"
)

var Module = fx.Options(
	fx.Provide(
		handler.NewProjectHandler,
		handler.NewEnvironmentHandler,
		handler.NewFeatureFlagHandler,
		handler.NewStrategyHandler,
		handler.NewAuthHandler,
		handler.NewRBACHandler,
		handler.NewTokenHandler,
		handler.NewAuditHandler,
		handler.NewClientHandler,
		fx.Annotate(
			NewRouter,
			fx.As(new(http.Handler)),
			fx.ResultTags(`name:"httpRouter"`),
		),
	),
)
