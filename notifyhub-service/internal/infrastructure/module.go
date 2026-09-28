package infrastructure

import (
	"time"

	httpadapter "github.com/JIeeiroSst/notifyhub-service/internal/adapter/primary/http"
	"github.com/JIeeiroSst/notifyhub-service/internal/adapter/secondary/channel"
	"github.com/JIeeiroSst/notifyhub-service/internal/adapter/secondary/fetcher"
	"github.com/JIeeiroSst/notifyhub-service/internal/adapter/secondary/metrics"
	"github.com/JIeeiroSst/notifyhub-service/internal/adapter/secondary/repository"
	"github.com/JIeeiroSst/notifyhub-service/internal/adapter/secondary/scheduler"
	"github.com/JIeeiroSst/notifyhub-service/internal/adapter/secondary/template"
	"github.com/JIeeiroSst/notifyhub-service/internal/application"
	"github.com/JIeeiroSst/notifyhub-service/internal/domain/port"
	"github.com/JIeeiroSst/notifyhub-service/internal/infrastructure/database"
	"github.com/JIeeiroSst/notifyhub-service/internal/infrastructure/server"
	"github.com/JIeeiroSst/notifyhub-service/internal/infrastructure/worker"
	"go.uber.org/fx"
)

const StopTimeout = 60 * time.Second

var Module = fx.Options(
	fx.Provide(newConfig),
	fx.Provide(newLogger),
	fx.Provide(newSettings),

	database.Module,
	repository.Module,

	channel.Module,
	scheduler.Module,
	fx.Provide(fetcher.NewFromConfig),
	fx.Provide(template.NewRenderer),
	fx.Provide(metrics.NewRecorder),
	fx.Provide(
		worker.NewPool,
		func(p *worker.Pool) port.TaskQueue { return p },
	),

	application.Module,

	httpadapter.Module,

	fx.Invoke(worker.RunPool),
	fx.Invoke(worker.RunScheduler),
	fx.Invoke(server.New),
)
