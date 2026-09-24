package eventbus

import (
	"github.com/JIeeiroSst/room-service/internal/domain/port"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(fx.Annotate(New, fx.As(new(port.EventPublisher)))),
)
