package infrastructure

import (
	httpadapter "github.com/JIeeiroSst/airflow-service/internal/adapter/primary/http"
	"github.com/JIeeiroSst/airflow-service/internal/adapter/secondary/airflowclient"
	"github.com/JIeeiroSst/airflow-service/internal/application"
	"github.com/JIeeiroSst/airflow-service/internal/infrastructure/server"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Invoke(initLogger),
	fx.Provide(newConfig),

	airflowclient.Module, // port.*Repository implementations backed by the Airflow REST API

	application.Module, // port.*Usecase implementations

	httpadapter.Module, // *httpadapter.Handler

	fx.Invoke(server.New),
)
