package infrastructure

import (
	"github.com/JIeeroSst/hospital-patient-management-service/config"
	"github.com/JIeeroSst/hospital-patient-management-service/internal/adapter/primary/auth"
	"github.com/JIeeroSst/hospital-patient-management-service/internal/adapter/primary/grpcapi"
	"github.com/JIeeroSst/hospital-patient-management-service/internal/adapter/primary/httpapi"
	"github.com/JIeeroSst/hospital-patient-management-service/internal/adapter/secondary/billing"
	"github.com/JIeeroSst/hospital-patient-management-service/internal/adapter/secondary/document"
	"github.com/JIeeroSst/hospital-patient-management-service/internal/adapter/secondary/ekyc"
	"github.com/JIeeroSst/hospital-patient-management-service/internal/adapter/secondary/notification"
	"github.com/JIeeroSst/hospital-patient-management-service/internal/adapter/secondary/repository"
	"github.com/JIeeroSst/hospital-patient-management-service/internal/adapter/secondary/userservice"
	"github.com/JIeeroSst/hospital-patient-management-service/internal/adapter/secondary/wallet"
	"github.com/JIeeroSst/hospital-patient-management-service/internal/application"
	"github.com/JIeeroSst/hospital-patient-management-service/internal/infrastructure/database"
	"github.com/JIeeroSst/hospital-patient-management-service/internal/infrastructure/server"
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

func initLogger() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.InfoLevel)
}

var Core = fx.Options(
	repository.Module,
	billing.Module,
	wallet.Module,
	notification.Module,
	ekyc.Module,
	document.Module,
	userservice.Module,
	application.Module,
	auth.Module,
	grpcapi.Module,
	httpapi.Module,
	fx.Invoke(server.New),
)

func validateConfig(cfg *config.Config) error { return cfg.Validate() }

var Module = fx.Options(
	fx.Invoke(initLogger),
	fx.Provide(config.Load),
	fx.Invoke(validateConfig),
	database.Module,
	Core,
)
