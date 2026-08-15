package minio

import (
	"go.uber.org/fx"

	"github.com/JIeeiroSst/photo-service/internal/application/ports"
)

var Module = fx.Module("storage.minio",
	fx.Provide(NewClient),
	fx.Provide(NewStorage),
	fx.Provide(func(s *Storage) ports.ImageStorage { return s }),
)
