package ingest

import (
	"context"
	"fmt"

	"github.com/joshk0/anyriichi/worker/internal/db"
	"github.com/joshk0/anyriichi/worker/internal/rules"
)

const replayBatch = 500

// ReplayRanks replays dan/pt for every queued player (rules.ReplayTenhou) and writes the results
// to game_participants, in batches until the queue is empty. Returns the number of players.
func (t *Tenhou) ReplayRanks(ctx context.Context) (int, error) {
	total := 0
	for {
		n, err := t.replayBatch(ctx)
		if err != nil {
			return total, err
		}
		total += n
		if n < replayBatch {
			if total > 0 {
				t.log.Info("rank replay done", "players", total)
			}
			return total, nil
		}
	}
}

// QueueReplayAll queues every Tenhou player for a replay (e.g. after a rules change).
func (t *Tenhou) QueueReplayAll(ctx context.Context) (int64, error) {
	return t.q.QueueRankReplayAll(ctx)
}

func (t *Tenhou) replayBatch(ctx context.Context) (int, error) {
	tx, err := t.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()
	qtx := t.q.WithTx(tx)
	ids, err := qtx.ClaimRankReplay(ctx, replayBatch)
	if err != nil || len(ids) == 0 {
		return 0, err
	}
	rows, err := qtx.ListReplayGames(ctx, ids)
	if err != nil {
		return 0, err
	}
	w := replayWrites(rows)
	if err := qtx.WriteReplayResults(ctx, w); err != nil {
		return 0, fmt.Errorf("write replay: %w", err)
	}
	return len(ids), tx.Commit(ctx)
}

// replayWrites groups rows (ordered by player, mode, time) into per-player-per-mode sequences,
// replays each, and returns the update arrays.
func replayWrites(rows []db.ListReplayGamesRow) db.WriteReplayResultsParams {
	var w db.WriteReplayResultsParams
	opt := func(p *int) int32 {
		if p == nil {
			return unknownDelta
		}
		return int32(*p)
	}
	for start := 0; start < len(rows); {
		end := start
		for end < len(rows) && rows[end].PlayerID == rows[start].PlayerID && rows[end].Players == rows[start].Players {
			end++
		}
		seq := rows[start:end]
		games := make([]rules.ReplayGame, len(seq))
		for i, r := range seq {
			games[i] = rules.ReplayGame{Players: int(r.Players), Level: r.Lobby, Length: r.Length, Placement: int(r.Placement)}
			if r.RankBefore.Valid {
				a := int(r.RankBefore.Int32)
				games[i].AnchorRank = &a
			}
		}
		steps, err := rules.ReplayTenhou(games)
		if err != nil { // bad data (e.g. unknown rank index): leave this sequence unknown
			steps = make([]rules.ReplayStep, len(seq))
		}
		for i, r := range seq {
			s := steps[i]
			w.GameIds = append(w.GameIds, r.GameID)
			w.Placements = append(w.Placements, r.Placement)
			w.RankAfter = append(w.RankAfter, opt(s.RankAfter))
			w.PtMin = append(w.PtMin, opt(s.PtMin))
			w.PtMax = append(w.PtMax, opt(s.PtMax))
			ev := ""
			if s.Event != nil {
				ev = string(*s.Event)
			}
			w.Events = append(w.Events, ev)
			w.Deltas = append(w.Deltas, opt(s.Delta))
		}
		start = end
	}
	return w
}
