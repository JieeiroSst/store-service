package clock

import (
	"time"

	"github.com/JIeeiroSst/auth-service/internal/port"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(func() port.Clock { return System{} }),
)

type System struct{}

func (System) Now() time.Time { return time.Now().UTC() }
