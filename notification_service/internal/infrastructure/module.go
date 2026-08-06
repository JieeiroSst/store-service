package infrastructure

import (
	httpadapter "github.com/JIeeiroSst/nofitifaction-service/internal/adapter/primary/http"
	workeradapter "github.com/JIeeiroSst/nofitifaction-service/internal/adapter/primary/worker"
	"github.com/JIeeiroSst/nofitifaction-service/internal/adapter/secondary/notifier"
	"github.com/JIeeiroSst/nofitifaction-service/internal/adapter/secondary/publisher"
	"github.com/JIeeiroSst/nofitifaction-service/internal/adapter/secondary/repository"
	"github.com/JIeeiroSst/nofitifaction-service/internal/adapter/secondary/slacktemplate"
	emailtemplate "github.com/JIeeiroSst/nofitifaction-service/internal/adapter/secondary/template"
	"github.com/JIeeiroSst/nofitifaction-service/internal/application"
	"github.com/JIeeiroSst/nofitifaction-service/internal/infrastructure/database"
	"github.com/JIeeiroSst/nofitifaction-service/internal/infrastructure/queue"
	"github.com/JIeeiroSst/nofitifaction-service/internal/infrastructure/server"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Invoke(initLogger),
	fx.Provide(newConfig),

	database.Module, // *gorm.DB
	queue.Module,    // rabbitmq.RabbitMQ

	repository.Module,    // port.*Repository
	notifier.Module,      // port.PushSender / EmailSender / SlackSender
	publisher.Module,     // port.NotificationPublisher
	emailtemplate.Module, // port.TemplateRenderer
	slacktemplate.Module, // port.SlackTemplateRenderer

	application.Module, // port.*Usecase

	httpadapter.Module,   // *httpadapter.Handler
	workeradapter.Module, // rabbitmq consumer lifecycle

	fx.Invoke(server.New),
)
