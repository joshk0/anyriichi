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
	"sync"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/joshk0/anyriichi/worker/internal/config"
	"github.com/joshk0/anyriichi/worker/internal/database"
	"github.com/joshk0/anyriichi/worker/internal/queue"
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

	// Without DATABASE_URL the worker still runs its scheduler (handy for local dev), but no queue.
	var pool *pgxpool.Pool
	if cfg.DatabaseURL != "" {
		pool, err = database.Open(ctx, cfg.DatabaseURL)
		if err != nil {
			log.Error("database", "err", err)
			os.Exit(1)
		}
		defer pool.Close()
	} else {
		log.Warn("DATABASE_URL not set; job queue disabled")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		if pool != nil {
			pctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
			defer cancel()
			if err := pool.Ping(pctx); err != nil {
				http.Error(w, "db: "+err.Error(), http.StatusServiceUnavailable)
				return
			}
		}
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

	var wg sync.WaitGroup
	if pool != nil {
		q := queue.New(pool, log, queue.Options{})
		q.Register(queue.Kind{
			Name: "system.ping",
			// Smoke-test job: SELECT queue.enqueue('system.ping', '{"from":"me"}');
			Handler: func(ctx context.Context, j queue.Job) error {
				log.Info("ping", "job_id", j.ID, "payload", string(j.Payload))
				return nil
			},
		})
		s.Add(scheduler.Job{Name: "queue-maintenance", Interval: 5 * time.Minute, RunAtStart: true, Run: q.Maintain})
		wg.Add(1)
		go func() { defer wg.Done(); q.Run(ctx) }()
	}

	log.Info("worker started", "http", cfg.HTTPAddr, "tenhou_poll", cfg.TenhouPollInterval, "db_configured", pool != nil)
	s.Run(ctx)
	wg.Wait()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
	log.Info("worker stopped")
}
