package application

import (
	"time"

	"github.com/JIeeiroSst/upload-service/internal/domain/port"
	"go.uber.org/fx"
)

type systemClock struct{}

func (systemClock) Now() time.Time { return time.Now().UTC() }

var Module = fx.Options(
	fx.Provide(
		func() port.Clock { return systemClock{} },
		NewFileService,
	),
)
