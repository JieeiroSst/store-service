package application

import (
	"go.uber.org/fx"

	"github.com/JIeeiroSst/toggle-service/internal/application/audit"
	"github.com/JIeeiroSst/toggle-service/internal/application/auth"
	"github.com/JIeeiroSst/toggle-service/internal/application/client"
	"github.com/JIeeiroSst/toggle-service/internal/application/environment"
	"github.com/JIeeiroSst/toggle-service/internal/application/evaluation"
	"github.com/JIeeiroSst/toggle-service/internal/application/featureflag"
	"github.com/JIeeiroSst/toggle-service/internal/application/project"
	"github.com/JIeeiroSst/toggle-service/internal/application/rbac"
	"github.com/JIeeiroSst/toggle-service/internal/application/strategy"
	"github.com/JIeeiroSst/toggle-service/internal/application/token"
)

var Module = fx.Options(
	fx.Provide(
		evaluation.NewEngine,
		audit.NewService,
		project.NewService,
		environment.NewService,
		featureflag.NewService,
		strategy.NewService,
		auth.NewService,
		rbac.NewService,
		token.NewService,
		client.NewService,
	),
)
