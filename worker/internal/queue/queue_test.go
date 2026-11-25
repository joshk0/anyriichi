package queue

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/joshk0/anyriichi/worker/internal/dbtest"
)

func discard() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }

func TestBackoff(t *testing.T) {
	for attempt, want := range map[int]time.Duration{0: 10 * time.Second, 1: 10 * time.Second, 2: 20 * time.Second, 4: 80 * time.Second, 20: time.Hour} {
		if got := Backoff(attempt); got != want {
			t.Errorf("Backoff(%d) = %v, want %v", attempt, got, want)
		}
	}
}

func TestEnqueueDedup(t *testing.T) {
	pool := dbtest.New(t)
	q := New(pool, discard(), Options{})
	ctx := context.Background()
	id1, err := q.Enqueue(ctx, "x", map[string]int{"a": 1}, EnqueueOptions{UniqueKey: "k"})
	if err != nil || id1 == 0 {
		t.Fatalf("first enqueue: id=%d err=%v", id1, err)
	}
	id2, err := q.Enqueue(ctx, "x", nil, EnqueueOptions{UniqueKey: "k"})
	if err != nil || id2 != 0 {
		t.Fatalf("duplicate enqueue: id=%d err=%v, want 0", id2, err)
	}
	if id3, _ := q.Enqueue(ctx, "y", nil, EnqueueOptions{UniqueKey: "k"}); id3 == 0 {
		t.Fatal("same key, different kind should enqueue")
	}
}

func TestRunProcessesRetriesAndFails(t *testing.T) {
	pool := dbtest.New(t)
	q := New(pool, discard(), Options{PollInterval: 50 * time.Millisecond})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var mu sync.Mutex
	seen := map[string]int{}
	done := make(chan struct{}, 10)
	q.Register(Kind{Name: "ok", Handler: func(_ context.Context, j Job) error {
		mu.Lock()
		seen[string(j.Payload)]++
		mu.Unlock()
		done <- struct{}{}
		return nil
	}})
	var flaky atomic.Int32
	q.Register(Kind{Name: "flaky", Handler: func(_ context.Context, j Job) error {
		if flaky.Add(1) == 1 {
			return errors.New("transient")
		}
		done <- struct{}{}
		return nil
	}})
	q.Register(Kind{Name: "bad", Handler: func(context.Context, Job) error {
		defer func() { done <- struct{}{} }()
		return Permanent(errors.New("nope"))
	}})

	go q.Run(ctx)
	time.Sleep(100 * time.Millisecond) // let LISTEN start

	okID, _ := q.Enqueue(ctx, "ok", map[string]string{"n": "1"}, EnqueueOptions{})
	flakyID, _ := q.Enqueue(ctx, "flaky", nil, EnqueueOptions{})
	badID, _ := q.Enqueue(ctx, "bad", nil, EnqueueOptions{})

	// ok + bad + first flaky failure, then force the flaky retry to be due now.
	wait(t, done, 2)
	waitState(t, q, flakyID, "pending")
	if _, err := pool.Exec(ctx, "update queue.jobs set run_at = now() where id = $1", flakyID); err != nil {
		t.Fatal(err)
	}
	wait(t, done, 1)

	waitState(t, q, okID, "succeeded")
	waitState(t, q, flakyID, "succeeded")
	waitState(t, q, badID, "failed")
	if j, _ := q.q.GetJob(ctx, flakyID); j.Attempts != 2 {
		t.Errorf("flaky attempts = %d, want 2", j.Attempts)
	}
	if j, _ := q.q.GetJob(ctx, badID); j.LastError == nil || *j.LastError != "nope" {
		t.Errorf("bad last_error = %v", j.LastError)
	}
	mu.Lock()
	defer mu.Unlock()
	if seen[`{"n": "1"}`] != 1 { // jsonb normalises spacing
		t.Errorf("ok payloads = %v", seen)
	}
}

func TestMaintainRescuesStaleJobs(t *testing.T) {
	pool := dbtest.New(t)
	q := New(pool, discard(), Options{StaleAfter: time.Minute, KeepFinished: time.Hour})
	ctx := context.Background()
	stale, _ := q.Enqueue(ctx, "x", nil, EnqueueOptions{})
	old, _ := q.Enqueue(ctx, "x", nil, EnqueueOptions{})
	if _, err := pool.Exec(ctx, `update queue.jobs set state='running', locked_at=now()-interval '2 minutes' where id=$1`, stale); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `update queue.jobs set state='succeeded', finished_at=now()-interval '2 hours' where id=$1`, old); err != nil {
		t.Fatal(err)
	}
	if err := q.Maintain(ctx); err != nil {
		t.Fatal(err)
	}
	waitState(t, q, stale, "pending")
	if _, err := q.q.GetJob(ctx, old); err == nil {
		t.Error("old succeeded job was not pruned")
	}
}

func TestConcurrentClaimsDoNotDuplicate(t *testing.T) {
	pool := dbtest.New(t)
	q := New(pool, discard(), Options{PollInterval: 50 * time.Millisecond})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	const n = 40
	var mu sync.Mutex
	counts := map[int64]int{}
	var total atomic.Int32
	q.Register(Kind{Name: "c", Concurrency: 4, Handler: func(_ context.Context, j Job) error {
		mu.Lock()
		counts[j.ID]++
		mu.Unlock()
		total.Add(1)
		return nil
	}})
	for i := 0; i < n; i++ {
		if _, err := q.Enqueue(ctx, "c", map[string]int{"i": i}, EnqueueOptions{}); err != nil {
			t.Fatal(err)
		}
	}
	go q.Run(ctx)
	deadline := time.Now().Add(10 * time.Second)
	for total.Load() < n && time.Now().Before(deadline) {
		time.Sleep(20 * time.Millisecond)
	}
	time.Sleep(100 * time.Millisecond)
	if len(counts) != n || total.Load() != n {
		t.Fatalf("processed %d distinct / %d total, want %d", len(counts), total.Load(), n)
	}
}

func wait(t *testing.T, ch <-chan struct{}, n int) {
	t.Helper()
	for i := 0; i < n; i++ {
		select {
		case <-ch:
		case <-time.After(10 * time.Second):
			t.Fatalf("timed out waiting for job %d/%d", i+1, n)
		}
	}
}

func waitState(t *testing.T, q *Queue, id int64, want string) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for {
		j, err := q.q.GetJob(context.Background(), id)
		if err == nil && j.State == want {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("job %d state = %q (err %v), want %q", id, j.State, err, want)
		}
		time.Sleep(20 * time.Millisecond)
	}
}
