// Package ingest turns platform data into rows: Tenhou public results (sc/raw) and full logs.
// Design: docs/design/worker/tenhou-ingest.md.
package ingest

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/joshk0/anyriichi/worker/internal/db"
	"github.com/joshk0/anyriichi/worker/internal/logref"
	"github.com/joshk0/anyriichi/worker/internal/queue"
	"github.com/joshk0/anyriichi/worker/internal/rules"
	"github.com/joshk0/anyriichi/worker/internal/tenhou"
)

// JobFetchTenhouLog is the queue kind for fetching one Tenhou full log. Payload: {"log_id": "…"}.
const JobFetchTenhouLog = "tenhou.log"

const (
	stateLastPoll  = "tenhou.raw.last_poll"
	unknownDelta   = -99999 // null marker for rank deltas (see migration)
	sourceScc      = "scc"
	sourceScb      = "scb"
	anyDanForDelta = 10 // non-last placement points don't depend on rank
	tenhouiRank    = 20 // 天鳳位 (rules/tenhou.json)
)

type TenhouConfig struct {
	// MinPollInterval is the minimum time between index polls, persisted across restarts.
	// Tenhou requires ≥ 20 minutes.
	MinPollInterval time.Duration
	// Backfill limits which hourly files are ingested: only those starting within this window.
	Backfill time.Duration
	// FetchHououLogs queues full logs for every 鳳凰 game seen in scc files.
	FetchHououLogs bool
}

type Tenhou struct {
	pool   *pgxpool.Pool
	q      *db.Queries
	client *tenhou.Client
	log    *slog.Logger
	cfg    TenhouConfig
	now    func() time.Time
}

func NewTenhou(pool *pgxpool.Pool, client *tenhou.Client, log *slog.Logger, cfg TenhouConfig) *Tenhou {
	return &Tenhou{pool: pool, q: db.New(pool), client: client, log: log.With("component", "tenhou"), cfg: cfg, now: time.Now}
}

type pollState struct {
	At time.Time `json:"at"`
}

// Poll fetches the sc/raw index and ingests new or grown hourly scb/scc files, oldest first.
// It enforces the minimum poll interval itself, so restarts can't poll too often.
func (t *Tenhou) Poll(ctx context.Context) error {
	if raw, err := t.q.GetWorkerState(ctx, stateLastPoll); err == nil {
		var st pollState
		if json.Unmarshal(raw, &st) == nil {
			if since := t.now().Sub(st.At); since < t.cfg.MinPollInterval {
				t.log.Info("tenhou poll skipped: too soon", "since_last", since.Round(time.Second))
				return nil
			}
		}
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("read poll state: %w", err)
	}
	// Record the attempt before fetching so failures also count toward the interval.
	st, _ := json.Marshal(pollState{At: t.now()})
	if err := t.q.SetWorkerState(ctx, db.SetWorkerStateParams{Key: stateLastPoll, Value: st}); err != nil {
		return fmt.Errorf("write poll state: %w", err)
	}

	files, err := t.client.Index(ctx)
	if err != nil {
		return fmt.Errorf("tenhou index: %w", err)
	}
	cutoff := t.now().Add(-t.cfg.Backfill)
	var candidates []tenhou.RawFile
	var names []string
	for _, f := range files {
		if _, hour, ok := tenhou.RankedHourly(f.Name); ok && !hour.Before(cutoff) {
			candidates = append(candidates, f)
			names = append(names, f.Name)
		}
	}
	known, err := t.q.GetRawFileSizes(ctx, names)
	if err != nil {
		return fmt.Errorf("raw file sizes: %w", err)
	}
	stored := make(map[string]int64, len(known))
	for _, k := range known {
		stored[k.File] = k.Size
	}
	var changed []tenhou.RawFile
	for _, f := range candidates {
		if f.Size > stored[f.Name] {
			changed = append(changed, f)
		}
	}
	tenhou.SortFiles(changed)
	t.log.Info("tenhou poll", "index", len(files), "in_window", len(candidates), "changed", len(changed))

	for _, f := range changed {
		if err := t.ingestFile(ctx, f); err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			var he *tenhou.HTTPError
			if errors.As(err, &he) && he.Status == http.StatusNotFound {
				t.log.Warn("tenhou raw file vanished", "file", f.Name)
				continue
			}
			return fmt.Errorf("ingest %s: %w", f.Name, err)
		}
	}
	return nil
}

func (t *Tenhou) ingestFile(ctx context.Context, f tenhou.RawFile) error {
	kind, hour, _ := tenhou.RankedHourly(f.Name)
	body, err := t.client.RawFile(ctx, f.Name)
	if err != nil {
		return err
	}
	n, skipped, queued, err := t.IngestResults(ctx, kind, hour, strings.HasSuffix(f.Name, ".html.gz"), body)
	if err != nil {
		return err
	}
	if err := t.q.UpsertRawFile(ctx, db.UpsertRawFileParams{File: f.Name, Size: f.Size, Lines: int32(n)}); err != nil {
		return err
	}
	t.log.Info("tenhou raw file ingested", "file", f.Name, "games", n, "skipped", skipped, "logs_queued", queued)
	return nil
}

// IngestResults upserts every game in one scb/scc file body (one transaction, one round trip per
// batch). It returns games ingested, lines skipped and full logs newly queued.
func (t *Tenhou) IngestResults(ctx context.Context, kind string, hour time.Time, htmlFile bool, body []byte) (n, skipped, queued int, err error) {
	var params []db.UpsertTenhouSummaryParams
	var logIDs []string
	sc := bufio.NewScanner(bytes.NewReader(body))
	sc.Buffer(make([]byte, 64<<10), 1<<20)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		res, perr := tenhou.ParseLine(line, hour, htmlFile)
		if perr != nil {
			skipped++
			t.log.Debug("tenhou line skipped", "err", perr)
			continue
		}
		params = append(params, summaryParams(res, kind))
		if res.LogID != "" && kind == sourceScc {
			logIDs = append(logIDs, res.LogID)
		}
	}
	if err := sc.Err(); err != nil {
		return 0, skipped, 0, err
	}

	tx, err := t.pool.Begin(ctx)
	if err != nil {
		return 0, skipped, 0, err
	}
	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()
	qtx := t.q.WithTx(tx)
	var batchErr error
	if len(params) > 0 {
		br := qtx.UpsertTenhouSummary(ctx, params)
		br.QueryRow(func(i int, _ int64, err error) {
			if err != nil && batchErr == nil {
				batchErr = fmt.Errorf("line %d: %w", i, err)
			}
		})
		if err := br.Close(); err != nil && batchErr == nil {
			batchErr = err
		}
	}
	if batchErr != nil {
		return 0, skipped, 0, batchErr
	}
	if t.cfg.FetchHououLogs {
		for _, id := range logIDs {
			ok, err := qtx.QueueTenhouLog(ctx, db.QueueTenhouLogParams{LogID: id, Source: sourceScc, Priority: -1})
			if err != nil {
				return 0, skipped, 0, fmt.Errorf("queue %s: %w", id, err)
			}
			if ok {
				queued++
			}
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, skipped, 0, err
	}
	return len(params), skipped, queued, nil
}

func summaryParams(r tenhou.Result, source string) db.UpsertTenhouSummaryParams {
	p := db.UpsertTenhouSummaryParams{
		MatchKey:  tenhou.MatchKey(r.StartedAt.Truncate(time.Hour), r.Rule.Code, r.Players),
		StartedAt: r.StartedAt,
		Duration:  int16(r.DurationMin),
		Source:    source,
		Code:      r.Rule.Code,
		Players:   int16(r.Rule.Players),
		Length:    r.Rule.Length,
		Lobby:     r.Rule.Lobby,
		Aka:       r.Rule.Aka,
		Kuitan:    r.Rule.Kuitan,
		Speed:     r.Rule.Speed,
	}
	for i, pl := range r.Players {
		p.Names = append(p.Names, pl.Name)
		p.PtTenths = append(p.PtTenths, int32(pl.PtTenths))
		delta := int32(unknownDelta)
		if i+1 < r.Rule.Players { // last place depends on the (unknown) rank
			if d, err := rules.TenhouRankDelta(anyDanForDelta, rules.TenhouGame{Players: r.Rule.Players, Level: r.Rule.Lobby, Length: r.Rule.Length, Placement: i + 1}); err == nil {
				delta = int32(d)
			}
		}
		p.RankDeltas = append(p.RankDeltas, delta)
	}
	return p
}

type fetchPayload struct {
	LogID string `json:"log_id"`
}

// QueueLog registers a log id for fetching (e.g. a user submission). Returns false if already known.
func (t *Tenhou) QueueLog(ctx context.Context, logID, source string) (bool, error) {
	return t.q.QueueTenhouLog(ctx, db.QueueTenhouLogParams{LogID: logID, Source: source})
}

// FetchLog is the queue handler for JobFetchTenhouLog.
func (t *Tenhou) FetchLog(ctx context.Context, job queue.Job) error {
	var p fetchPayload
	if err := json.Unmarshal(job.Payload, &p); err != nil || p.LogID == "" {
		return queue.Permanent(fmt.Errorf("bad payload %s", job.Payload))
	}
	log := t.log.With("log_id", p.LogID)

	ref := logref.Parse(p.LogID)
	switch {
	case ref == nil || ref.Platform != "tenhou" || ref.ID != p.LogID:
		return t.reject(ctx, p.LogID, "invalid_log_id")
	case !ref.Ranked:
		return t.reject(ctx, p.LogID, "not_ranked")
	case ref.Rule != nil && ref.Rule.Shuugi:
		return t.reject(ctx, p.LogID, "not_ranked")
	case !rules.TenhouLogSupported(p.LogID):
		return t.reject(ctx, p.LogID, "before_rule_change") // D-046
	}

	body, err := t.client.Log(ctx, p.LogID)
	if err != nil {
		var he *tenhou.HTTPError
		if errors.As(err, &he) && he.Status == http.StatusNotFound {
			return t.fail(ctx, p.LogID, "not_found", queue.Permanent(err))
		}
		return err // transient: retry with backoff
	}
	l, err := tenhou.ParseMjlogBytes(body)
	if err != nil {
		return t.fail(ctx, p.LogID, "parse_error", queue.Permanent(err))
	}
	if l.Lobby != 0 {
		return t.reject(ctx, p.LogID, "not_ranked")
	}
	if l.RuleCode != ref.RuleCode {
		return t.fail(ctx, p.LogID, "rule_mismatch", queue.Permanent(fmt.Errorf("log id rule %s, GO type %s", ref.RuleCode, l.RuleCode)))
	}
	gameID, err := t.storeLog(ctx, p.LogID, l, body)
	if err != nil {
		return err
	}
	log.Info("tenhou log stored", "game_id", gameID, "hands", len(l.Hands))
	return nil
}

func (t *Tenhou) storeLog(ctx context.Context, logID string, l *tenhou.Log, raw []byte) (int64, error) {
	rule, err := tenhou.DecodeRuleCode(l.RuleCode)
	if err != nil {
		return 0, queue.Permanent(err)
	}
	hour, err := time.ParseInLocation("2006010215", logID[:10], tenhou.JST)
	if err != nil {
		return 0, queue.Permanent(err)
	}
	p := db.UpsertTenhouFullLogParams{
		MatchKey: tenhou.MatchKey(hour, rule.Code, l.ByPlacement()), StartedAt: hour,
		Code: rule.Code, Players: int16(rule.Players), Length: rule.Length, Lobby: rule.Lobby,
		Aka: rule.Aka, Kuitan: rule.Kuitan, Speed: rule.Speed,
	}
	bySeat := make([]int, l.Players) // placement → seat
	for s, seat := range l.Seats {
		bySeat[seat.Placement-1] = s
	}
	for place, s := range bySeat {
		seat := l.Seats[s]
		p.Names = append(p.Names, seat.Name)
		p.PtTenths = append(p.PtTenths, int32(seat.PtTenths))
		p.Seats = append(p.Seats, int32(s))
		p.Scores = append(p.Scores, int32(seat.FinalScore))
		p.Dans = append(p.Dans, int32(seat.Dan))
		p.Rates = append(p.Rates, seat.Rate)
		delta := int32(unknownDelta)
		if d, err := rules.TenhouRankDelta(seat.Dan, rules.TenhouGame{Players: rule.Players, Level: rule.Lobby, Length: rule.Length, Placement: place + 1}); err == nil {
			delta = int32(d)
		}
		if seat.Dan == tenhouiRank { // 天鳳位: no pt changes
			delta = 0
		}
		p.RankDeltas = append(p.RankDeltas, delta)
	}

	var gz bytes.Buffer
	zw := gzip.NewWriter(&gz)
	_, _ = zw.Write(raw)
	_ = zw.Close()

	tx, err := t.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()
	qtx := t.q.WithTx(tx)
	gameID, err := qtx.UpsertTenhouFullLog(ctx, p)
	if err != nil {
		return 0, fmt.Errorf("upsert game: %w", err)
	}
	if err := qtx.DeleteHands(ctx, gameID); err != nil {
		return 0, err
	}
	for hi, h := range l.Hands {
		if err := qtx.InsertHand(ctx, db.InsertHandParams{GameID: gameID, HandIndex: int16(hi), Round: int16(h.Round),
			Honba: int16(h.Honba), RiichiSticks: int16(h.RiichiSticks), DealerSeat: int16(h.Dealer), Outcome: h.Outcome}); err != nil {
			return 0, fmt.Errorf("hand %d: %w", hi, err)
		}
		for s, hp := range h.Players {
			if err := qtx.InsertHandPlayerStats(ctx, handPlayerParams(gameID, hi, s, hp)); err != nil {
				return 0, fmt.Errorf("hand %d seat %d: %w", hi, s, err)
			}
		}
	}
	if err := qtx.MarkTenhouLog(ctx, db.MarkTenhouLogParams{LogID: logID, Status: "stored", RawGz: gz.Bytes(),
		GameID: pgtype.Int8{Int64: gameID, Valid: true}}); err != nil {
		return 0, err
	}
	return gameID, tx.Commit(ctx)
}

func handPlayerParams(gameID int64, hand, seat int, hp tenhou.HandPlayer) db.InsertHandPlayerStatsParams {
	p := db.InsertHandPlayerStatsParams{
		GameID: gameID, HandIndex: int16(hand), Seat: int16(seat),
		Won: hp.Won, DealtIn: hp.DealtIn, Riichi: hp.Riichi, Called: hp.Called,
		Yaku: []int16{}, ScoreDelta: int32(hp.ScoreDelta),
	}
	if hp.WonBy != "" {
		w := hp.WonBy
		p.WonBy = &w
	}
	if hp.TenpaiAtDraw != nil {
		p.TenpaiAtDraw = pgtype.Bool{Bool: *hp.TenpaiAtDraw, Valid: true}
	}
	if hp.Won {
		p.Han = pgtype.Int2{Int16: int16(hp.Han), Valid: true}
		p.Fu = pgtype.Int2{Int16: int16(hp.Fu), Valid: true}
		p.WinValue = pgtype.Int4{Int32: int32(hp.WinValue), Valid: true}
		for _, y := range hp.Yaku {
			p.Yaku = append(p.Yaku, int16(y))
		}
	}
	if hp.DealtIn {
		p.DealInValue = pgtype.Int4{Int32: int32(hp.DealInValue), Valid: true}
	}
	return p
}

func (t *Tenhou) reject(ctx context.Context, logID, reason string) error {
	t.log.Info("tenhou log rejected", "log_id", logID, "reason", reason)
	return t.q.MarkTenhouLog(context.WithoutCancel(ctx), db.MarkTenhouLogParams{LogID: logID, Status: "rejected", RejectReason: &reason})
}

func (t *Tenhou) fail(ctx context.Context, logID, reason string, err error) error {
	if merr := t.q.MarkTenhouLog(context.WithoutCancel(ctx), db.MarkTenhouLogParams{LogID: logID, Status: "failed", RejectReason: &reason}); merr != nil {
		return merr
	}
	return err
}

// Reparse re-derives games and per-hand rows from stored raw logs (no network), e.g. after a
// parser change. Returns the number of logs reparsed.
func (t *Tenhou) Reparse(ctx context.Context) (int, error) {
	logs, err := t.q.ListStoredTenhouLogs(ctx)
	if err != nil {
		return 0, err
	}
	for _, l := range logs {
		zr, err := gzip.NewReader(bytes.NewReader(l.RawGz))
		if err != nil {
			return 0, fmt.Errorf("%s: %w", l.LogID, err)
		}
		raw, err := io.ReadAll(zr)
		if err != nil {
			return 0, fmt.Errorf("%s: %w", l.LogID, err)
		}
		parsed, err := tenhou.ParseMjlogBytes(raw)
		if err != nil {
			return 0, fmt.Errorf("%s: %w", l.LogID, err)
		}
		if _, err := t.storeLog(ctx, l.LogID, parsed, raw); err != nil {
			return 0, fmt.Errorf("%s: %w", l.LogID, err)
		}
	}
	return len(logs), nil
}
