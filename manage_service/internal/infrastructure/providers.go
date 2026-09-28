package infrastructure

import (
	"os"

	"github.com/JIeeiroSst/manage-service/config"
	"github.com/JIeeiroSst/manage-service/internal/application"
	"github.com/JIeeiroSst/manage-service/pkg/consul"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func newConfig() (*config.Config, error) {
	dir, err := config.ReadFileEnv(".env")
	if err != nil {
		return nil, err
	}

	var cfg *config.Config
	if os.Getenv("NODE_ENV") != "" {
		cfg, err = consul.NewConfigConsul(dir.HostConsul, dir.KeyConsul, dir.ServiceConsul).ConnectConfigConsul()
	} else {
		cfg, err = config.ReadConf(config.Path())
	}
	if err != nil {
		return nil, err
	}
	cfg.ApplyEnv()
	return cfg, nil
}

func newLogger() (*zap.Logger, error) {
	zc := zap.NewProductionConfig()
	zc.OutputPaths = []string{"stdout"}
	zc.ErrorOutputPaths = []string{"stderr"}
	log, err := zc.Build()
	if err != nil {
		return nil, err
	}
	return log.With(zap.String("service", "manage-service")), nil
}

func FxLogger(log *zap.Logger) fxevent.Logger {
	return &fxevent.ZapLogger{Logger: log.WithOptions(zap.IncreaseLevel(zapcore.WarnLevel))}
}

func newAdminCredentials(cfg *config.Config) application.AdminCredentials {
	return application.AdminCredentials{
		User:     cfg.Keycloak.AdminUser,
		Password: cfg.Keycloak.AdminPassword,
		Realm:    cfg.Keycloak.AdminRealm,
	}
}
