// Command worker is anyRiichi's ingestion worker: polls platform data sources, fetches and parses
// game logs, and maintains statistics. See docs/design/worker/architecture.md.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joshk0/anyriichi/worker/internal/config"
	"github.com/joshk0/anyriichi/worker/internal/scheduler"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config", "err", err)
		os.Exit(1)
	}
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel}))
	slog.SetDefault(log)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})
	srv := &http.Server{Addr: cfg.HTTPAddr, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("http server", "err", err)
			stop()
		}
	}()

	s := scheduler.New(log)
	s.Add(scheduler.Job{
		Name:     "tenhou-raw-poll",
		Interval: cfg.TenhouPollInterval,
		Run: func(ctx context.Context) error {
			// Implemented in M1: docs/design/worker/tenhou-ingest.md
			log.Debug("tenhou-raw-poll not implemented yet")
			return nil
		},
	})

	log.Info("worker started", "http", cfg.HTTPAddr, "tenhou_poll", cfg.TenhouPollInterval, "db_configured", cfg.DatabaseURL != "")
	s.Run(ctx)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
	log.Info("worker stopped")
}
