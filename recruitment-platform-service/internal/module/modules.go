package module

import (
	"context"
	"fmt"

	aiadapter "github.com/JIeeiroSst/recruitment-platform-service/internal/adapter/ai"
	"github.com/JIeeiroSst/recruitment-platform-service/internal/adapter/eventbus"
	"github.com/JIeeiroSst/recruitment-platform-service/internal/adapter/http/handler"
	"github.com/JIeeiroSst/recruitment-platform-service/internal/adapter/http/middleware"
	notifadapter "github.com/JIeeiroSst/recruitment-platform-service/internal/adapter/notification"
	"github.com/JIeeiroSst/recruitment-platform-service/internal/adapter/persistence/postgres"
	temporaladapter "github.com/JIeeiroSst/recruitment-platform-service/internal/adapter/temporal"
	"github.com/JIeeiroSst/recruitment-platform-service/internal/adapter/temporal/activity"
	domainapp "github.com/JIeeiroSst/recruitment-platform-service/internal/domain/application"
	"github.com/JIeeiroSst/recruitment-platform-service/internal/domain/candidate"
	"github.com/JIeeiroSst/recruitment-platform-service/internal/domain/job"
	"github.com/JIeeiroSst/recruitment-platform-service/internal/domain/referral"
	"github.com/JIeeiroSst/recruitment-platform-service/internal/port"
	applicationusecase "github.com/JIeeiroSst/recruitment-platform-service/internal/usecase/application"
	candidateusecase "github.com/JIeeiroSst/recruitment-platform-service/internal/usecase/candidate"
	jobusecase "github.com/JIeeiroSst/recruitment-platform-service/internal/usecase/job"
	referralusecase "github.com/JIeeiroSst/recruitment-platform-service/internal/usecase/referral"
	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"go.temporal.io/sdk/client"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var LoggerModule = fx.Module("logger",
	fx.Provide(func(cfg *Config) (*zap.Logger, error) {
		if cfg.App.Env == "production" {
			return zap.NewProduction()
		}
		return zap.NewDevelopment()
	}),
)

var DatabaseModule = fx.Module("database",
	fx.Provide(
		NewDatabase,
		fx.Annotate(postgres.NewCandidateRepository, fx.As(new(candidate.Repository))),
		fx.Annotate(postgres.NewJobRepository, fx.As(new(job.Repository))),
		fx.Annotate(postgres.NewApplicationRepository, fx.As(new(domainapp.Repository))),
		fx.Annotate(postgres.NewPartnerRepository, fx.As(new(referral.PartnerRepository))),
		fx.Annotate(postgres.NewReferralRepository, fx.As(new(referral.ReferralRepository))),
		fx.Annotate(postgres.NewPayoutRepository, fx.As(new(referral.PayoutRepository))),
	),
)

func NewDatabase(cfg *Config, logger *zap.Logger) (*sqlx.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.Name,
		cfg.Database.User, cfg.Database.Password, cfg.Database.SSLMode,
	)
	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("database: connect failed: %w", err)
	}
	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	logger.Info("database connected",
		zap.String("host", cfg.Database.Host),
		zap.String("dbname", cfg.Database.Name),
	)
	return db, nil
}

var InfraModule = fx.Module("infra",
	fx.Provide(
		func(cfg *Config, repo candidate.Repository, log *zap.Logger) port.AIService {
			return aiadapter.NewOpenAIService(aiadapter.Config{
				BaseURL: cfg.AI.BaseURL,
				APIKey:  cfg.AI.APIKey,
				Model:   cfg.AI.Model,
			}, repo, log)
		},
		func(cfg *Config, log *zap.Logger) port.NotificationService {
			return notifadapter.NewSendGridService(notifadapter.Config{
				APIKey:    cfg.Notification.APIKey,
				FromEmail: cfg.Notification.FromEmail,
				FromName:  cfg.Notification.FromName,
			}, log)
		},
		fx.Annotate(eventbus.New, fx.As(new(port.EventBus))),
	),
)

var TemporalModule = fx.Module("temporal",
	fx.Provide(
		NewTemporalClient,
		activity.NewActivities,
		temporaladapter.NewWorker,
		fx.Annotate(temporaladapter.NewWorkflowService, fx.As(new(port.WorkflowService))),
	),
	fx.Invoke(registerTemporalWorker),
)

func NewTemporalClient(cfg *Config, logger *zap.Logger) (client.Client, error) {
	c, err := client.Dial(client.Options{
		HostPort:  cfg.Temporal.HostPort,
		Namespace: cfg.Temporal.Namespace,
	})
	if err != nil {
		return nil, fmt.Errorf("temporal: dial failed: %w", err)
	}
	logger.Info("temporal connected",
		zap.String("host", cfg.Temporal.HostPort),
		zap.String("namespace", cfg.Temporal.Namespace),
	)
	return c, nil
}

func registerTemporalWorker(lc fx.Lifecycle, w *temporaladapter.Worker, logger *zap.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("starting temporal worker")
			return w.Start()
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("stopping temporal worker")
			w.Stop()
			return nil
		},
	})
}

var ServiceModule = fx.Module("service",
	fx.Provide(
		fx.Annotate(candidateusecase.NewService, fx.As(new(port.CandidateService))),
		fx.Annotate(jobusecase.NewService, fx.As(new(port.JobService))),
		fx.Annotate(applicationusecase.NewService, fx.As(new(port.ApplicationService))),
		fx.Annotate(referralusecase.NewService, fx.As(new(port.ReferralService))),
	),
)

var HTTPModule = fx.Module("http",
	fx.Provide(
		handler.NewCandidateHandler,
		handler.NewJobHandler,
		handler.NewApplicationHandler,
		handler.NewReferralHandler,
		handler.NewAnalyticsHandler,
		NewRouter,
	),
	fx.Invoke(registerHTTPServer),
)

func NewRouter(
	cfg *Config,
	logger *zap.Logger,
	candidateH *handler.CandidateHandler,
	jobH *handler.JobHandler,
	appH *handler.ApplicationHandler,
	referralH *handler.ReferralHandler,
	analyticsH *handler.AnalyticsHandler,
) *gin.Engine {
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestID())
	r.Use(zapMiddleware(logger))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "env": cfg.App.Env})
	})

	r.GET("/api/v1/referrals/track/:token", referralH.TrackClick)

	api := r.Group("/api/v1", middleware.JWTAuth(cfg.App.JWTSecret, logger))
	candidateH.RegisterRoutes(api)
	jobH.RegisterRoutes(api)
	appH.RegisterRoutes(api)
	referralH.RegisterRoutesProtected(api)
	analyticsH.RegisterRoutes(api)

	return r
}

func registerHTTPServer(lc fx.Lifecycle, r *gin.Engine, cfg *Config, logger *zap.Logger) {
	addr := fmt.Sprintf(":%d", cfg.App.Port)
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("HTTP server starting", zap.String("addr", addr))
			go func() {
				if err := r.Run(addr); err != nil {
					logger.Fatal("HTTP server error", zap.Error(err))
				}
			}()
			return nil
		},
	})
}

func zapMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		logger.Info("http",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.String("request_id", c.GetString("request_id")),
		)
	}
}
