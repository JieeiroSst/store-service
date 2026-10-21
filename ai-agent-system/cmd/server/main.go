package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"aiagentsystem/internal/adapter/primary/httpapi"
	"aiagentsystem/internal/adapter/primary/ollamaapi"
	"aiagentsystem/internal/adapter/secondary/mysql"
	"aiagentsystem/internal/adapter/secondary/ollama"
	"aiagentsystem/internal/application"
	"aiagentsystem/internal/config"
	"aiagentsystem/internal/infrastructure/database"
	"aiagentsystem/internal/infrastructure/server"
)

const startupTimeout = 10 * time.Second

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	if err := run(logger); err != nil {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	startupCtx, cancelStartup := context.WithTimeout(ctx, startupTimeout)
	defer cancelStartup()

	db, err := database.OpenMySQL(startupCtx, cfg.MySQLDSN)
	if err != nil {
		return err
	}
	defer db.Close()

	historyRepo := mysql.NewHistoryRepository(db)
	if err := historyRepo.Migrate(startupCtx); err != nil {
		return err
	}

	llm := ollama.New(ollama.Config{BaseURL: cfg.OllamaBaseURL, Model: cfg.OllamaModel})
	chat := application.NewChatService(llm, historyRepo, logger)
	defer chat.Wait()
	history := application.NewHistoryService(historyRepo)

	mux := http.NewServeMux()
	mux.Handle("/", httpapi.NewRouter(httpapi.Dependencies{Chat: chat, History: history, Ready: db.PingContext}))
	mux.Handle("/api/", ollamaapi.NewRouter(ollamaapi.NewHandlers(chat, cfg.OllamaModel)))

	logger.Info("server listening",
		"port", cfg.Port,
		"version", config.Version,
		"ollama_base_url", cfg.OllamaBaseURL,
		"ollama_model", cfg.OllamaModel,
	)
	return server.New(":"+cfg.Port, mux, logger).Run(ctx)
}
