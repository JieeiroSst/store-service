package main

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"aiagentsystem/internal/chat"
	"aiagentsystem/internal/config"
	"aiagentsystem/internal/faq"
	"aiagentsystem/internal/history"
	"aiagentsystem/internal/httpapi"
	"aiagentsystem/internal/ollamaapi"
	"aiagentsystem/internal/proxy"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	proxyClient := proxy.New(proxy.Config{
		OllamaBaseURL: cfg.OllamaBaseURL,
		OllamaModel:   cfg.OllamaModel,
		VoyageAPIKey:  cfg.VoyageAPIKey,
	})

	faqEntries, err := faq.LoadFromFile(cfg.FAQFile)
	if err != nil {
		logger.Error("failed to load FAQ file", "error", err)
		os.Exit(1)
	}

	faqStore := faq.NewStore(cfg.FAQSimilarityThreshold)

	loadCtx, cancelLoad := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancelLoad()
	if err := faqStore.Load(loadCtx, proxyClient.Voyage, faqEntries); err != nil {
		// FAQ matching is a stated requirement across every entry point, not
		// a best-effort feature, so a failure to embed the FAQ set at
		// startup should fail the server rather than silently degrade.
		logger.Error("failed to load FAQ embeddings", "error", err)
		os.Exit(1)
	}
	logger.Info("loaded FAQ entries", "count", len(faqEntries))

	db, err := sql.Open("mysql", cfg.MySQLDSN)
	if err != nil {
		logger.Error("failed to open MySQL connection", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	pingCtx, cancelPing := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelPing()
	if err := db.PingContext(pingCtx); err != nil {
		logger.Error("failed to connect to MySQL", "error", err)
		os.Exit(1)
	}

	historyStore := history.New(db)
	schemaCtx, cancelSchema := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelSchema()
	if err := historyStore.EnsureSchema(schemaCtx); err != nil {
		logger.Error("failed to ensure chat history schema", "error", err)
		os.Exit(1)
	}

	service := chat.NewService(proxyClient.Ollama, proxyClient.Voyage, faqStore, historyStore, logger)

	mux := http.NewServeMux()
	mux.Handle("/v1/", httpapi.NewRouter(service, historyStore, logger))
	mux.Handle("/api/", ollamaapi.NewRouter(ollamaapi.NewHandlers(service, cfg.OllamaModel)))

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: mux,
	}

	serverErr := make(chan error, 1)
	go func() {
		logger.Info("server listening", "port", cfg.Port, "ollama_base_url", cfg.OllamaBaseURL, "ollama_model", cfg.OllamaModel)
		serverErr <- srv.ListenAndServe()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	case sig := <-stop:
		logger.Info("shutting down", "signal", sig.String())

		shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancelShutdown()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Error("graceful shutdown failed", "error", err)
			os.Exit(1)
		}
	}
}
