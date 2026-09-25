package application

import (
	"github.com/JIeeiroSst/movie-recommendation-service/internal/domain/port"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(NewService,
			fx.As(new(port.RecommendationUsecase)),
			fx.As(new(port.ModelMaintainer)),
		),
	),
)
