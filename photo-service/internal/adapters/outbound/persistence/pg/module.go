package pg

import (
	"go.uber.org/fx"

	"github.com/JIeeiroSst/photo-service/internal/application/ports"
)

var Module = fx.Module("postgres",
	fx.Provide(NewPool),
	fx.Provide(NewJobRepository),
	fx.Provide(func(r *JobRepository) ports.JobRepository { return r }),
	fx.Invoke(RegisterMigrations),
)
