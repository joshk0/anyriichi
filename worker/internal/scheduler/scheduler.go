// Package scheduler runs named jobs at fixed intervals. Each job runs at most once at a time.
// A Postgres-backed queue (river or SKIP LOCKED) will handle on-demand work in M1; see docs/design/worker/architecture.md.
package scheduler

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

type Job struct {
	Name     string
	Interval time.Duration
	// RunAtStart runs the job immediately instead of waiting one interval.
	RunAtStart bool
	Run        func(ctx context.Context) error
}

type Scheduler struct {
	log  *slog.Logger
	jobs []Job
}

func New(log *slog.Logger) *Scheduler { return &Scheduler{log: log} }

func (s *Scheduler) Add(j Job) { s.jobs = append(s.jobs, j) }

// Run blocks until ctx is cancelled and all in-flight jobs return.
func (s *Scheduler) Run(ctx context.Context) {
	var wg sync.WaitGroup
	for _, j := range s.jobs {
		wg.Add(1)
		go func(j Job) {
			defer wg.Done()
			t := time.NewTicker(j.Interval)
			defer t.Stop()
			if j.RunAtStart {
				s.runOnce(ctx, j)
			}
			for {
				select {
				case <-ctx.Done():
					return
				case <-t.C:
					s.runOnce(ctx, j)
				}
			}
		}(j)
	}
	wg.Wait()
}

func (s *Scheduler) runOnce(ctx context.Context, j Job) {
	start := time.Now()
	err := j.Run(ctx)
	attrs := []any{"job", j.Name, "duration", time.Since(start)}
	if err != nil {
		s.log.Error("job failed", append(attrs, "err", err)...)
		return
	}
	s.log.Info("job finished", attrs...)
}
