package probe

import (
	"github.com/JIeeiroSst/networking-service/internal/port"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(func() port.Prober { return NewProber() }),
)
