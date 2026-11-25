// Package queue is the worker's Postgres-backed job queue (table queue.jobs, D-043).
//
// Jobs are claimed with FOR UPDATE SKIP LOCKED. Each registered kind runs its own fixed number of
// claim loops, which doubles as per-host politeness (e.g. Concurrency 1 for tenhou.net). Enqueue
// sends NOTIFY queue_jobs; workers LISTEN for low latency and also poll as a fallback.
package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"os"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/joshk0/anyriichi/worker/internal/db"
)

const notifyChannel = "queue_jobs"

// Job is a claimed job handed to a Handler.
type Job struct {
	ID          int64
	Kind        string
	Payload     json.RawMessage
	Attempt     int // 1-based
	MaxAttempts int
}

// Handler processes one job. Returning nil completes it; returning an error retries it with
// exponential backoff until MaxAttempts; wrap the error with Permanent to fail it immediately.
type Handler func(ctx context.Context, job Job) error

// Kind registers a handler for a job kind.
type Kind struct {
	Name        string
	Concurrency int           // claim loops for this kind; default 1
	Timeout     time.Duration // per-job timeout; default 5m (must stay below StaleAfter)
	Handler     Handler
}

type permanentError struct{ err error }

func (e permanentError) Error() string { return e.err.Error() }
func (e permanentError) Unwrap() error { return e.err }

// Permanent marks an error as non-retryable.
func Permanent(err error) error { return permanentError{err} }

// Options tune the queue. Zero values pick defaults.
type Options struct {
	PollInterval time.Duration // fallback poll when no NOTIFY arrives; default 30s
	StaleAfter   time.Duration // running jobs older than this are rescued; default 15m
	KeepFinished time.Duration // succeeded jobs are pruned after this; default 7 days
	WorkerID     string        // default hostname:pid
}

type Queue struct {
	pool  *pgxpool.Pool
	q     *db.Queries
	log   *slog.Logger
	opts  Options
	kinds map[string]Kind
	wake  map[string]chan struct{}
}

func New(pool *pgxpool.Pool, log *slog.Logger, opts Options) *Queue {
	if opts.PollInterval == 0 {
		opts.PollInterval = 30 * time.Second
	}
	if opts.StaleAfter == 0 {
		opts.StaleAfter = 15 * time.Minute
	}
	if opts.KeepFinished == 0 {
		opts.KeepFinished = 7 * 24 * time.Hour
	}
	if opts.WorkerID == "" {
		host, _ := os.Hostname()
		opts.WorkerID = fmt.Sprintf("%s:%d", host, os.Getpid())
	}
	return &Queue{pool: pool, q: db.New(pool), log: log, opts: opts,
		kinds: map[string]Kind{}, wake: map[string]chan struct{}{}}
}

// Register adds a handler. Call before Run.
func (q *Queue) Register(k Kind) {
	if k.Concurrency <= 0 {
		k.Concurrency = 1
	}
	if k.Timeout <= 0 {
		k.Timeout = 5 * time.Minute
	}
	q.kinds[k.Name] = k
	q.wake[k.Name] = make(chan struct{}, k.Concurrency)
}

// EnqueueOptions are optional enqueue parameters.
type EnqueueOptions struct {
	RunAt       time.Time // default now
	Priority    int16
	UniqueKey   string // de-duplicates against pending/running jobs of the same kind
	MaxAttempts int32  // default 5
}

// Enqueue inserts a job. It returns id 0 when UniqueKey matched an active job.
func (q *Queue) Enqueue(ctx context.Context, kind string, payload any, o EnqueueOptions) (int64, error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return 0, fmt.Errorf("marshal payload: %w", err)
	}
	if payload == nil {
		b = []byte("{}")
	}
	if o.RunAt.IsZero() {
		o.RunAt = time.Now()
	}
	if o.MaxAttempts <= 0 {
		o.MaxAttempts = 5
	}
	var uk *string
	if o.UniqueKey != "" {
		uk = &o.UniqueKey
	}
	return q.q.Enqueue(ctx, db.EnqueueParams{Kind: kind, Payload: b, RunAt: o.RunAt,
		Priority: o.Priority, UniqueKey: uk, MaxAttempts: o.MaxAttempts})
}

// Run starts the listener and claim loops and blocks until ctx is cancelled and in-flight jobs end.
func (q *Queue) Run(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() { defer wg.Done(); q.listen(ctx) }()
	for _, k := range q.kinds {
		for i := 0; i < k.Concurrency; i++ {
			wg.Add(1)
			go func() { defer wg.Done(); q.loop(ctx, k) }()
		}
	}
	q.log.Info("queue started", "worker_id", q.opts.WorkerID, "kinds", len(q.kinds))
	wg.Wait()
}

func (q *Queue) loop(ctx context.Context, k Kind) {
	timer := time.NewTimer(0)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
		case <-q.wake[k.Name]:
		}
		// Drain: keep claiming until the kind has no ready jobs.
		for ctx.Err() == nil {
			worked, err := q.work(ctx, k)
			if err != nil {
				q.log.Error("queue claim", "kind", k.Name, "err", err)
				break
			}
			if !worked {
				break
			}
		}
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
		timer.Reset(jitter(q.opts.PollInterval))
	}
}

// work claims and runs at most one job. It reports whether a job was found.
func (q *Queue) work(ctx context.Context, k Kind) (bool, error) {
	row, err := q.q.ClaimJob(ctx, db.ClaimJobParams{LockedBy: q.opts.WorkerID, Kind: k.Name})
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	job := Job{ID: row.ID, Kind: row.Kind, Payload: row.Payload, Attempt: int(row.Attempts), MaxAttempts: int(row.MaxAttempts)}
	log := q.log.With("job_id", job.ID, "kind", job.Kind, "attempt", job.Attempt)

	start := time.Now()
	runErr := q.call(ctx, k, job)

	// Record the outcome even if ctx was cancelled during shutdown.
	dbCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancel()
	if runErr == nil {
		if err := q.q.CompleteJob(dbCtx, job.ID); err != nil {
			return true, fmt.Errorf("complete job %d: %w", job.ID, err)
		}
		log.Info("job succeeded", "duration", time.Since(start))
		return true, nil
	}
	var perm permanentError
	permanent := errors.As(runErr, &perm)
	backoff := Backoff(job.Attempt)
	if ctx.Err() != nil && !permanent {
		backoff = 0 // shutting down: hand the job straight back
	}
	if err := q.q.FailJob(dbCtx, db.FailJobParams{ID: job.ID, Permanent: permanent,
		BackoffSeconds: backoff.Seconds(), LastError: truncate(runErr.Error(), 2000)}); err != nil {
		return true, fmt.Errorf("fail job %d: %w", job.ID, err)
	}
	log.Warn("job failed", "err", runErr, "permanent", permanent, "retry_in", backoff, "duration", time.Since(start))
	return true, nil
}

func (q *Queue) call(ctx context.Context, k Kind, job Job) (err error) {
	ctx, cancel := context.WithTimeout(ctx, k.Timeout)
	defer cancel()
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	return k.Handler(ctx, job)
}

// listen holds one connection with LISTEN queue_jobs and wakes the matching claim loops.
func (q *Queue) listen(ctx context.Context) {
	for ctx.Err() == nil {
		if err := q.listenOnce(ctx); err != nil && ctx.Err() == nil {
			q.log.Warn("queue listener disconnected; polling until it reconnects", "err", err)
			select {
			case <-ctx.Done():
			case <-time.After(5 * time.Second):
			}
		}
	}
}

func (q *Queue) listenOnce(ctx context.Context) error {
	pc, err := q.pool.Acquire(ctx)
	if err != nil {
		return err
	}
	// Take the connection out of the pool: it stays in LISTEN state until closed.
	conn := pc.Hijack()
	defer func() { _ = conn.Close(context.Background()) }()
	if _, err := conn.Exec(ctx, "LISTEN "+notifyChannel); err != nil {
		return err
	}
	for {
		n, err := conn.WaitForNotification(ctx)
		if err != nil {
			return err
		}
		if ch, ok := q.wake[n.Payload]; ok {
			select {
			case ch <- struct{}{}:
			default:
			}
		}
	}
}

// Maintain rescues jobs abandoned by a crashed worker and prunes old succeeded jobs.
// Run it periodically from the scheduler.
func (q *Queue) Maintain(ctx context.Context) error {
	rescued, err := q.q.RescueStaleJobs(ctx, q.opts.StaleAfter.Seconds())
	if err != nil {
		return fmt.Errorf("rescue stale jobs: %w", err)
	}
	pruned, err := q.q.PruneFinishedJobs(ctx, q.opts.KeepFinished.Seconds())
	if err != nil {
		return fmt.Errorf("prune jobs: %w", err)
	}
	if rescued > 0 || pruned > 0 {
		q.log.Info("queue maintenance", "rescued", rescued, "pruned", pruned)
	}
	return nil
}

// Backoff is the retry delay after the given (1-based) attempt: 10s, 20s, 40s… capped at 1h.
func Backoff(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	d := 10 * time.Second
	for i := 1; i < attempt && d < time.Hour; i++ {
		d *= 2
	}
	return min(d, time.Hour)
}

func jitter(d time.Duration) time.Duration {
	return d - d/10 + time.Duration(rand.Int64N(int64(d/5)+1))
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
