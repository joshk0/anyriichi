package scheduler

import (
	"context"
	"io"
	"log/slog"
	"sync/atomic"
	"testing"
	"time"
)

func TestRunsAtStartAndOnInterval(t *testing.T) {
	var n atomic.Int32
	s := New(slog.New(slog.NewTextHandler(io.Discard, nil)))
	s.Add(Job{Name: "t", Interval: 20 * time.Millisecond, RunAtStart: true, Run: func(context.Context) error {
		n.Add(1)
		return nil
	}})
	ctx, cancel := context.WithTimeout(context.Background(), 70*time.Millisecond)
	defer cancel()
	s.Run(ctx)
	if got := n.Load(); got < 3 {
		t.Fatalf("ran %d times, want >= 3", got)
	}
}
