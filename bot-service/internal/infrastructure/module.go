package infrastructure

import (
	httpadapter "github.com/JIeeiroSst/bot-service/internal/adapter/primary/http"
	schedulerPrimary "github.com/JIeeiroSst/bot-service/internal/adapter/primary/scheduler"
	telegramPrimary "github.com/JIeeiroSst/bot-service/internal/adapter/primary/telegram"
	twitterPrimary "github.com/JIeeiroSst/bot-service/internal/adapter/primary/twitter"
	facebookSecondary "github.com/JIeeiroSst/bot-service/internal/adapter/secondary/facebook"
	repositorySecondary "github.com/JIeeiroSst/bot-service/internal/adapter/secondary/repository"
	routerSecondary "github.com/JIeeiroSst/bot-service/internal/adapter/secondary/router"
	telegramSecondary "github.com/JIeeiroSst/bot-service/internal/adapter/secondary/telegram"
	twitterSecondary "github.com/JIeeiroSst/bot-service/internal/adapter/secondary/twitter"
	"github.com/JIeeiroSst/bot-service/internal/application"
	"github.com/JIeeiroSst/bot-service/internal/infrastructure/database"
	"github.com/JIeeiroSst/bot-service/internal/infrastructure/server"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Invoke(initLogger),
	fx.Provide(newConfig),

	database.Module,            // *gorm.DB
	repositorySecondary.Module, // port.PostRepository

	telegramSecondary.Module, // *tgbotapi.BotAPI, port.ChannelSender + port.ChannelPublisher (telegram)
	twitterSecondary.Module,  // *twitter.Client, port.ChannelSender + port.ChannelPublisher (twitter)
	facebookSecondary.Module, // *facebook.Client, port.ChannelSender + port.ChannelPublisher (facebook)
	routerSecondary.Module,   // port.MessageSender, fans out by model.Channel

	application.Module, // port.BotUsecase, port.ContentUsecase, application.PublisherRegistry

	telegramPrimary.Module,  // *telegramadapter.Poller
	twitterPrimary.Module,   // *twitterprimary.Poller
	schedulerPrimary.Module, // *scheduler.Scheduler
	httpadapter.Module,      // *httpadapter.Handler, *httpadapter.FacebookHandler, *httpadapter.ContentHandler

	fx.Invoke(server.NewTelegramServer),
	fx.Invoke(server.NewTwitterServer),
	fx.Invoke(server.NewSchedulerServer),
	fx.Invoke(server.NewHTTPServer),
)
