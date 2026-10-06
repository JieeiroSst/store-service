package metrics

import (
	"github.com/JIeeiroSst/serpapi-service/internal/port"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(NewPrometheus),
	fx.Provide(func(p *Prometheus) port.Metrics { return p }),
)
