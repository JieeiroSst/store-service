package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/JIeeiroSst/notifyhub-service/internal/api"
	"github.com/JIeeiroSst/notifyhub-service/internal/api/handler"
	"github.com/JIeeiroSst/notifyhub-service/internal/channel"
	emailSender "github.com/JIeeiroSst/notifyhub-service/internal/channel/email"
	fbSender "github.com/JIeeiroSst/notifyhub-service/internal/channel/firebase"
	smsSender "github.com/JIeeiroSst/notifyhub-service/internal/channel/sms"
	"github.com/JIeeiroSst/notifyhub-service/internal/config"
	"github.com/JIeeiroSst/notifyhub-service/internal/fetcher"
	jobSvc "github.com/JIeeiroSst/notifyhub-service/internal/job"
	"github.com/JIeeiroSst/notifyhub-service/internal/pkg/logger"
	"github.com/JIeeiroSst/notifyhub-service/internal/scheduler"
	"github.com/JIeeiroSst/notifyhub-service/internal/store/mysql"
	tmplEngine "github.com/JIeeiroSst/notifyhub-service/internal/template"
	"github.com/JIeeiroSst/notifyhub-service/internal/worker"
	"go.uber.org/zap"
)

func main() {
	// ── Logger ────────────────────────────────────────────────────────────────
	log := logger.Must("production")
	defer log.Sync() //nolint:errcheck

	log.Info("notifyhub-service starting", zap.String("pid", fmt.Sprintf("%d", os.Getpid())))

	// ── Config ────────────────────────────────────────────────────────────────
	cfgPath := ".env"
	if v := os.Getenv("CONFIG_PATH"); v != "" {
		cfgPath = v
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatal("load config", zap.Error(err))
	}
	log.Info("config loaded", zap.String("port", cfg.Server.Port))

	// ── Database ──────────────────────────────────────────────────────────────
	store, err := mysql.New(
		cfg.Database.DSN,
		cfg.Database.MaxOpenConns,
		cfg.Database.MaxIdleConns,
		cfg.Database.ConnMaxLifetime,
	)
	if err != nil {
		log.Fatal("connect mysql", zap.Error(err))
	}
	if err := store.Ping(); err != nil {
		log.Fatal("mysql ping", zap.Error(err))
	}
	log.Info("mysql connected",
		zap.Int("max_open", cfg.Database.MaxOpenConns),
		zap.Int("max_idle", cfg.Database.MaxIdleConns),
	)

	// ── Channel Registry ──────────────────────────────────────────────────────
	registry := channel.NewRegistry()

	// Email channel
	registry.Register(emailSender.New(cfg.Email, log))
	log.Info("email channel registered", zap.String("provider", cfg.Email.Provider))

	// SMS channel
	if cfg.SMS.TwilioSID != "" || cfg.SMS.VonageKey != "" {
		registry.Register(smsSender.New(cfg.SMS, log))
		log.Info("sms channel registered", zap.String("provider", cfg.SMS.Provider))
	}

	// Firebase channel (optional)
	ctx := context.Background()
	if cfg.Firebase.CredentialsFile != "" {
		fb, err := fbSender.New(ctx, cfg.Firebase, log)
		if err != nil {
			log.Warn("firebase init failed — channel disabled", zap.Error(err))
		} else {
			registry.Register(fb)
			log.Info("firebase channel registered", zap.String("project", cfg.Firebase.ProjectID))
		}
	}

	// ── Template Engine ───────────────────────────────────────────────────────
	tmpl := tmplEngine.NewEngine()

	// Pre-compile all active templates at startup
	templates, err := store.ListTemplates(ctx, "")
	if err != nil {
		log.Warn("load templates for pre-compilation", zap.Error(err))
	} else {
		compiled := 0
		for _, t := range templates {
			if compileErr := tmpl.Compile(t); compileErr != nil {
				log.Warn("template compile error",
					zap.String("id", t.ID),
					zap.String("name", t.Name),
					zap.Error(compileErr),
				)
			} else {
				compiled++
			}
		}
		log.Info("templates pre-compiled", zap.Int("count", compiled))
	}

	// ── Fetcher ───────────────────────────────────────────────────────────────
	f := fetcher.New(cfg.Worker.FetcherPoolSize, cfg.Worker.FetchTimeoutSec)
	log.Info("http fetcher created",
		zap.Int("pool_size", cfg.Worker.FetcherPoolSize),
		zap.Int("timeout_sec", cfg.Worker.FetchTimeoutSec),
	)

	// ── Worker Pool ───────────────────────────────────────────────────────────
	pool := worker.NewPool(
		cfg.Worker.PoolSize,
		cfg.Worker.QueueSize,
		cfg.Worker.RetryMax,
		cfg.Worker.RetryDelay,
		registry,
		store,
		f,
		tmpl,
		log,
	)
	log.Info("worker pool started",
		zap.Int("goroutines", cfg.Worker.PoolSize),
		zap.Int("queue_size", cfg.Worker.QueueSize),
		zap.Int("retry_max", cfg.Worker.RetryMax),
	)

	// ── Scheduler ─────────────────────────────────────────────────────────────
	sched, err := scheduler.New(store, pool, log)
	if err != nil {
		log.Fatal("create scheduler", zap.Error(err))
	}
	if err := sched.Start(ctx); err != nil {
		log.Fatal("start scheduler", zap.Error(err))
	}

	// ── Service Layer ─────────────────────────────────────────────────────────
	js := jobSvc.NewService(store, sched, tmpl, log)
	cs := jobSvc.NewChannelService(store)
	ds := jobSvc.NewDataSourceService(store)
	ts := jobSvc.NewTemplateService(store, tmpl)

	// ── HTTP Server ───────────────────────────────────────────────────────────
	h := handler.New(store, js, cs, ds, ts, log)
	router := api.NewRouter(h, sched, cfg.Server.JWTSecret, cfg.Server.RateLimit, log)

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:           router,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1 MB
	}

	// ── Graceful Shutdown ─────────────────────────────────────────────────────
	serverErr := make(chan error, 1)
	go func() {
		log.Info("http server listening", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		log.Error("server error", zap.Error(err))
	case sig := <-quit:
		log.Info("shutdown signal received", zap.String("signal", sig.String()))
	}

	log.Info("shutting down gracefully...")

	shutCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 1. Stop accepting new HTTP requests
	if err := srv.Shutdown(shutCtx); err != nil {
		log.Error("http server shutdown", zap.Error(err))
	}

	// 2. Stop scheduler from submitting new tasks
	if err := sched.Stop(); err != nil {
		log.Error("scheduler shutdown", zap.Error(err))
	}

	// 3. Drain worker pool (wait for in-flight notifications)
	pool.Stop()

	log.Info("shutdown complete")
}
