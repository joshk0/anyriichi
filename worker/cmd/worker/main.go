// Command worker is anyRiichi's ingestion worker: polls platform data sources, fetches and parses
// game logs, and maintains statistics. See docs/design/worker/architecture.md.
package main

import (
	"context"
	"errors"
	"fmt"
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
	"github.com/joshk0/anyriichi/worker/internal/ingest"
	"github.com/joshk0/anyriichi/worker/internal/logref"
	"github.com/joshk0/anyriichi/worker/internal/queue"
	"github.com/joshk0/anyriichi/worker/internal/scheduler"
	"github.com/joshk0/anyriichi/worker/internal/tenhou"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config", "err", err)
		os.Exit(1)
	}
	if len(os.Args) > 1 {
		os.Exit(runCommand(cfg, os.Args[1:]))
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
	if pool == nil {
		log.Warn("tenhou ingestion disabled: no database")
	}

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

		th := ingest.NewTenhou(pool, tenhou.NewClient(cfg.UserAgent), log, ingest.TenhouConfig{
			MinPollInterval: cfg.TenhouPollInterval,
			Backfill:        cfg.TenhouBackfill,
			FetchHououLogs:  cfg.TenhouFetchHououLogs,
		})
		// One full-log fetch at a time: tenhou.net allows a single session (the client also serialises).
		q.Register(queue.Kind{Name: ingest.JobFetchTenhouLog, Concurrency: 1, Timeout: 2 * time.Minute, Handler: th.FetchLog})
		// Poll persists its last run in the DB, so RunAtStart can't exceed the 20-minute minimum.
		s.Add(scheduler.Job{Name: "tenhou-raw-poll", Interval: cfg.TenhouPollInterval, RunAtStart: true, Run: th.Poll})
		// Dan/pt replay for players whose games changed (queued by ingest).
		s.Add(scheduler.Job{Name: "tenhou-rank-replay", Interval: time.Minute, RunAtStart: true,
			Run: func(ctx context.Context) error { _, err := th.ReplayRanks(ctx); return err }})
		wg.Add(1)
		go func() { defer wg.Done(); q.Run(ctx) }()
	}

	log.Info("worker started", "http", cfg.HTTPAddr, "tenhou_poll", cfg.TenhouPollInterval, "db_configured", pool != nil)
	s.Run(ctx)
	<-ctx.Done() // the scheduler has no jobs without a database
	wg.Wait()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
	log.Info("worker stopped")
}

// runCommand handles one-off CLI commands:
//
//	worker enqueue <tenhou log id or URL>…   queue full-log fetches (as a submission)
//	worker reparse                           re-derive rows from stored raw logs (no network)
//	worker replay                            re-run the dan/pt replay for every player
func runCommand(cfg config.Config, args []string) int {
	oneArg := len(args) == 1 && (args[0] == "reparse" || args[0] == "replay")
	if !(args[0] == "enqueue" && len(args) >= 2) && !oneArg {
		fmt.Fprintln(os.Stderr, "usage: worker enqueue <tenhou log id or URL>... | worker reparse | worker replay")
		return 2
	}
	ctx := context.Background()
	pool, err := database.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		fmt.Fprintln(os.Stderr, "database:", err)
		return 1
	}
	defer pool.Close()
	th := ingest.NewTenhou(pool, tenhou.NewClient(cfg.UserAgent), slog.Default(), ingest.TenhouConfig{})
	if args[0] == "reparse" {
		n, err := th.Reparse(ctx)
		if err != nil {
			fmt.Fprintln(os.Stderr, "reparse:", err)
			return 1
		}
		fmt.Printf("reparsed %d logs\n", n)
		return 0
	}
	if args[0] == "replay" {
		if _, err := th.QueueReplayAll(ctx); err != nil {
			fmt.Fprintln(os.Stderr, "replay:", err)
			return 1
		}
		n, err := th.ReplayRanks(ctx)
		if err != nil {
			fmt.Fprintln(os.Stderr, "replay:", err)
			return 1
		}
		fmt.Printf("replayed %d players\n", n)
		return 0
	}
	code := 0
	for _, a := range args[1:] {
		ref := logref.Parse(a)
		if ref == nil || ref.Platform != "tenhou" {
			fmt.Fprintf(os.Stderr, "%s: not a Tenhou log\n", a)
			code = 1
			continue
		}
		queued, err := th.QueueLog(ctx, ref.ID, "submission")
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", ref.ID, err)
			code = 1
			continue
		}
		fmt.Printf("%s queued=%v\n", ref.ID, queued)
	}
	return code
}
