package videoclient

import (
	"github.com/JIeeiroSst/movie-recommendation-service/internal/domain/port"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(fx.Annotate(New, fx.As(new(port.CatalogSource)))),
)
