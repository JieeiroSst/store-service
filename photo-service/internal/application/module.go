package application

import (
	"go.uber.org/fx"

	"github.com/JIeeiroSst/photo-service/internal/application/ports"
	"github.com/JIeeiroSst/photo-service/internal/application/service"
)

var Module = fx.Module("application",
	fx.Provide(service.NewComposeService),
	fx.Provide(func(s *service.ComposeService) ports.ComposeImageUseCase { return s }),
)
