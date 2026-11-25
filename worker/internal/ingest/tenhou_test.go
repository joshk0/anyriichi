package ingest

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/joshk0/anyriichi/worker/internal/dbtest"
	"github.com/joshk0/anyriichi/worker/internal/queue"
	"github.com/joshk0/anyriichi/worker/internal/tenhou"
)

// Synthetic data only (public repo; Tenhou forbids redistribution).
const (
	scbBody = "03:07 | 20 | 四鳳南喰赤－ | Alice(+46.7) ボブ(+5.5) Dave(-17.5) Carol(-34.7)\n" +
		"03:10 | 14 | 四般東喰赤－ | Eve(+44.5) NoName(+7.6) Frank(-17.3) Alice(-34.8)\n" +
		"03:11 | 14 | 四若東喰赤－祝５ | X(+1.0) Y(+0.0) Z(-0.5) W(-0.5)\n"
	sccBody = `03:07 | 20 | 四鳳南喰赤－ | <a href="http://tenhou.net/0/?log=2026112603gm-00a9-0000-0123abcd">牌譜</a> | Alice(+46.7) ボブ(+5.5) Dave(-17.5) Carol(-34.7)<br>` + "\n"
	logID   = "2026112603gm-00a9-0000-0123abcd"
)

func gz(s string) []byte {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	_, _ = w.Write([]byte(s))
	_ = w.Close()
	return b.Bytes()
}

type fakeTenhou struct {
	srv      *httptest.Server
	requests atomic.Int32
}

func newFakeTenhou(t *testing.T) *fakeTenhou {
	mjlog, err := os.ReadFile("../tenhou/testdata/game4p.mjlog")
	if err != nil {
		t.Fatal(err)
	}
	f := &fakeTenhou{}
	f.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.requests.Add(1)
		switch {
		case r.URL.Path == "/sc/raw/list.cgi":
			fmt.Fprintf(w, "list([\n{file:'sca20261125.log.gz',size:10},\n{file:'scb2026112603.log.gz',size:%d},\n{file:'scc2026112603.html.gz',size:%d}\n]);", len(gz(scbBody)), len(gz(sccBody)))
		case r.URL.Path == "/sc/raw/dat/scb2026112603.log.gz":
			_, _ = w.Write(gz(scbBody))
		case r.URL.Path == "/sc/raw/dat/scc2026112603.html.gz":
			_, _ = w.Write(gz(sccBody))
		case r.URL.Path == "/0/log/" && r.URL.RawQuery == logID:
			_, _ = w.Write(mjlog)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(f.srv.Close)
	return f
}

func setup(t *testing.T) (*Tenhou, *fakeTenhou) {
	pool := dbtest.New(t)
	fake := newFakeTenhou(t)
	c := tenhou.NewClient("test")
	c.Base = fake.srv.URL
	c.MinGap = 0
	ing := NewTenhou(pool, c, slog.New(slog.NewTextHandler(io.Discard, nil)), TenhouConfig{
		MinPollInterval: 20 * time.Minute, Backfill: 24 * time.Hour, FetchHououLogs: true,
	})
	ing.now = func() time.Time { return time.Date(2026, 11, 26, 5, 0, 0, 0, tenhou.JST) }
	return ing, fake
}

func TestPollIngestsResultsAndQueuesHououLogs(t *testing.T) {
	ing, fake := setup(t)
	ctx := context.Background()
	if err := ing.Poll(ctx); err != nil {
		t.Fatal(err)
	}
	var games, participants, players int
	_ = ing.pool.QueryRow(ctx, "select count(*) from games").Scan(&games)
	_ = ing.pool.QueryRow(ctx, "select count(*) from game_participants").Scan(&participants)
	_ = ing.pool.QueryRow(ctx, "select count(*) from players").Scan(&players)
	if games != 2 || participants != 8 || players != 6 { // shuugi line skipped; NoName not a player
		t.Errorf("games=%d participants=%d players=%d", games, participants, players)
	}
	var sources []string
	var minute int
	_ = ing.pool.QueryRow(ctx, "select sources, extract(minute from started_at)::int from games g join rulesets r on r.id = g.ruleset_id where r.code = '00a9'").Scan(&sources, &minute)
	if strings.Join(sources, ",") != "scb,scc" || minute != 7 {
		t.Errorf("houou game sources=%v minute=%d", sources, minute)
	}
	var deltas []*int32
	rows, _ := ing.pool.Query(ctx, "select rank_pt_delta from game_participants gp join games g on g.id = gp.game_id join rulesets r on r.id = g.ruleset_id where r.code = '00a9' order by placement")
	for rows.Next() {
		var d *int32
		_ = rows.Scan(&d)
		deltas = append(deltas, d)
	}
	if len(deltas) != 4 || *deltas[0] != 90 || *deltas[1] != 45 || *deltas[2] != 0 || deltas[3] != nil {
		t.Errorf("summary rank deltas = %v", deltas)
	}
	var queued, jobs int
	_ = ing.pool.QueryRow(ctx, "select count(*) from private.tenhou_logs where status = 'queued'").Scan(&queued)
	_ = ing.pool.QueryRow(ctx, "select count(*) from queue.jobs where kind = 'tenhou.log'").Scan(&jobs)
	if queued != 1 || jobs != 1 {
		t.Errorf("queued=%d jobs=%d", queued, jobs)
	}

	// A second poll inside 20 minutes must not touch tenhou.net.
	before := fake.requests.Load()
	if err := ing.Poll(ctx); err != nil {
		t.Fatal(err)
	}
	if fake.requests.Load() != before {
		t.Error("polled again within MinPollInterval")
	}
	// After the interval, unchanged files are not downloaded again.
	ing.now = func() time.Time { return time.Date(2026, 11, 26, 5, 30, 0, 0, tenhou.JST) }
	if err := ing.Poll(ctx); err != nil {
		t.Fatal(err)
	}
	if got := fake.requests.Load() - before; got != 1 {
		t.Errorf("second poll made %d requests, want 1 (index only)", got)
	}
}

func TestFetchLogStoresFullLogAndMergesWithSummary(t *testing.T) {
	ing, _ := setup(t)
	ctx := context.Background()
	if err := ing.Poll(ctx); err != nil {
		t.Fatal(err)
	}
	payload, _ := json.Marshal(fetchPayload{LogID: logID})
	if err := ing.FetchLog(ctx, queue.Job{Payload: payload}); err != nil {
		t.Fatal(err)
	}
	var games int
	_ = ing.pool.QueryRow(ctx, "select count(*) from games").Scan(&games)
	if games != 2 {
		t.Errorf("full log should merge into the summary game; games = %d", games)
	}
	var sources []string
	var full bool
	var minute int
	_ = ing.pool.QueryRow(ctx, "select sources, has_full_log, extract(minute from started_at)::int from games g join rulesets r on r.id = g.ruleset_id where r.code = '00a9'").Scan(&sources, &full, &minute)
	if strings.Join(sources, ",") != "log,scb,scc" || !full || minute != 7 {
		t.Errorf("sources=%v full=%v minute=%d", sources, full, minute)
	}
	type row struct {
		name  string
		seat  int
		score int
		rank  int
		delta *int32
	}
	var got []row
	rows, _ := ing.pool.Query(ctx, `select name, seat, final_score, rank_before, rank_pt_delta from game_participants gp
		join games g on g.id = gp.game_id join rulesets r on r.id = g.ruleset_id where r.code = '00a9' order by placement`)
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.name, &r.seat, &r.score, &r.rank, &r.delta); err != nil {
			t.Fatal(err)
		}
		got = append(got, r)
	}
	if len(got) != 4 || got[0].name != "Alice" || got[0].seat != 0 || got[0].score != 33700 || *got[0].delta != 90 ||
		got[3].name != "Carol" || got[3].rank != 20 || *got[3].delta != 0 || // 天鳳位: no pt
		got[2].name != "Dave" || *got[2].delta != 0 {
		t.Errorf("participants: %+v", got)
	}
	var hands, stats, riichi int
	_ = ing.pool.QueryRow(ctx, "select count(*) from private.hands").Scan(&hands)
	_ = ing.pool.QueryRow(ctx, "select count(*), count(*) filter (where riichi) from private.hand_player_stats where player_id is not null").Scan(&stats, &riichi)
	if hands != 3 || stats != 12 || riichi != 1 {
		t.Errorf("hands=%d stats=%d riichi=%d", hands, stats, riichi)
	}
	var status string
	var raw []byte
	_ = ing.pool.QueryRow(ctx, "select status, raw_gz from private.tenhou_logs where log_id = $1", logID).Scan(&status, &raw)
	if status != "stored" || len(raw) == 0 {
		t.Errorf("status=%s raw=%d bytes", status, len(raw))
	}
	// Re-fetching is idempotent.
	if err := ing.FetchLog(ctx, queue.Job{Payload: payload}); err != nil {
		t.Fatal(err)
	}
	_ = ing.pool.QueryRow(ctx, "select count(*) from private.hands").Scan(&hands)
	if hands != 3 {
		t.Errorf("hands after refetch = %d", hands)
	}
}

func TestFetchLogRejects(t *testing.T) {
	ing, _ := setup(t)
	ctx := context.Background()
	for id, reason := range map[string]string{
		"2016010100gm-00a9-0000-0123abcd": "before_rule_change",
		"2026112603gm-00a9-0001-0123abcd": "not_ranked",
	} {
		if _, err := ing.QueueLog(ctx, id, "submission"); err != nil {
			t.Fatal(err)
		}
		payload, _ := json.Marshal(fetchPayload{LogID: id})
		if err := ing.FetchLog(ctx, queue.Job{Payload: payload}); err != nil {
			t.Fatalf("%s: %v", id, err)
		}
		var status, got string
		_ = ing.pool.QueryRow(ctx, "select status, reject_reason from private.tenhou_logs where log_id = $1", id).Scan(&status, &got)
		if status != "rejected" || got != reason {
			t.Errorf("%s: status=%s reason=%s, want rejected/%s", id, status, got, reason)
		}
	}
}
