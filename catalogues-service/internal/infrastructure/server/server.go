package server

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/JIeeiroSst/catalogues-service/config"
	"github.com/JIeeiroSst/catalogues-service/internal/adapter/primary/httpapi"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

type Params struct {
	fx.In

	Lifecycle fx.Lifecycle
	Config    *config.Config
	DB        *gorm.DB
	Handlers  []httpapi.Handler `group:"handlers"`
}

// New mounts every route group and ties the HTTP server to the fx lifecycle.
func New(p Params) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", health(p.DB))
	for _, h := range p.Handlers {
		h.Register(mux)
	}

	srv := &http.Server{
		Addr:              ":" + p.Config.Server.HTTPPort,
		Handler:           recoverHTTP(logRequests(mux)),
		ReadHeaderTimeout: 10 * time.Second,
	}

	p.Lifecycle.Append(fx.Hook{
		OnStart: func(context.Context) error {
			ln, err := net.Listen("tcp", srv.Addr)
			if err != nil {
				return err
			}
			go func() {
				if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
					slog.Error("http server stopped", "error", err)
				}
			}()
			slog.Info("http listening", "addr", ln.Addr().String())
			return nil
		},
		OnStop: func(stopCtx context.Context) error {
			ctx, done := context.WithTimeout(stopCtx, p.Config.Server.ShutdownTimeout)
			defer done()
			return srv.Shutdown(ctx)
		},
	})
}

func health(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		sqlDB, err := db.DB()
		if err == nil {
			err = sqlDB.PingContext(ctx)
		}
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"status":"unavailable"}`))
			return
		}
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)
		if r.URL.Path == "/health" {
			return
		}
		slog.Info("http request", "method", r.Method, "path", r.URL.Path,
			"status", sw.status, "duration", time.Since(start).String())
	})
}

func recoverHTTP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if v := recover(); v != nil {
				slog.Error("panic serving request", "method", r.Method, "path", r.URL.Path,
					"panic", v, "stack", string(debug.Stack()))
				http.Error(w, "internal error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
