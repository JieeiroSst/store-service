package storage

import (
	"github.com/JIeeiroSst/cdn-service/internal/ports"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(NewMinIOStorage),
	fx.Provide(func(s *MinIOStorage) ports.ObjectStorage { return s }),
)
