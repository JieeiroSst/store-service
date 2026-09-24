package infrastructure

import (
	"github.com/JIeeiroSst/customer-relationship-service/config"
	httpadapter "github.com/JIeeiroSst/customer-relationship-service/internal/adapter/primary/http"
	"github.com/JIeeiroSst/customer-relationship-service/internal/adapter/primary/scheduler"
	"github.com/JIeeiroSst/customer-relationship-service/internal/adapter/primary/temporalworker"
	"github.com/JIeeiroSst/customer-relationship-service/internal/adapter/secondary/antivirus"
	"github.com/JIeeiroSst/customer-relationship-service/internal/adapter/secondary/digisign"
	"github.com/JIeeiroSst/customer-relationship-service/internal/adapter/secondary/notifier"
	"github.com/JIeeiroSst/customer-relationship-service/internal/adapter/secondary/objectstore"
	"github.com/JIeeiroSst/customer-relationship-service/internal/adapter/secondary/orchestrator"
	"github.com/JIeeiroSst/customer-relationship-service/internal/adapter/secondary/pdfdoc"
	"github.com/JIeeiroSst/customer-relationship-service/internal/adapter/secondary/repository"
	"github.com/JIeeiroSst/customer-relationship-service/internal/adapter/secondary/sealer"
	"github.com/JIeeiroSst/customer-relationship-service/internal/application"
	"github.com/JIeeiroSst/customer-relationship-service/internal/infrastructure/database"
	"github.com/JIeeiroSst/customer-relationship-service/internal/infrastructure/server"
	"github.com/JIeeiroSst/customer-relationship-service/internal/infrastructure/temporal"
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

func initLogger() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.InfoLevel)
}

var Core = fx.Options(
	repository.Module,
	notifier.Module,
	digisign.Module,
	pdfdoc.Module,
	objectstore.Module,
	antivirus.Module,
	sealer.Module,
	temporal.Module,
	orchestrator.Module,
	application.Module,
	httpadapter.Module,
	scheduler.Module,
	temporalworker.Module,
	fx.Invoke(server.New),
)

var Module = fx.Options(
	fx.Invoke(initLogger),
	fx.Provide(config.Load),
	database.Module,
	Core,
)
