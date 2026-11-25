package rules

// Tenhou dan/pt replay (docs/design/worker/tenhou-ingest.md "Dan / pt replay").
//
// Tenhou never publishes a player's rank points, and full logs only tell us the rank (dan) at the
// start of a game. So the replay tracks a small set of hypotheses — (rank, pt range) — through a
// player's games in order:
//
//   - Before the first anchor (a game whose full log shows the player's rank) nothing is known.
//   - At an anchor, hypotheses with a different rank are dropped. If none survive (first anchor, or
//     data we can't reconcile, e.g. games outside our window), the player restarts at that rank
//     with pt anywhere in [0, promotePt).
//   - Each game moves every hypothesis by the game's rank-point change. A range that straddles a
//     promotion or demotion threshold splits into "stayed" and "promoted/demoted" hypotheses; a
//     later anchor tells which one happened, and the earlier game's event is then filled in.
//   - A promotion or demotion resets pt to the new rank's start, so once one is pinned down the pt
//     is known exactly from then on.
//
// For players with a full log of every game (鳳凰) this is exact after their first promotion or
// demotion; for others the pt range narrows as submitted logs arrive.

// ReplayGame is one ranked game of one player, in play order, for one mode (3p and 4p ranks are
// separate, so replay them separately).
type ReplayGame struct {
	Players   int
	Level     string // ippan | joukyuu | tokujou | houou
	Length    string // east | south
	Placement int
	// AnchorRank is the player's rank at the start of the game, when a full log shows it.
	AnchorRank *int
}

// ReplayStep is what the replay knows about one game. Fields are nil when unknown or ambiguous.
type ReplayStep struct {
	RankBefore *int
	Delta      *int
	RankAfter  *int
	PtMin      *int // pt range after the game (rank RankAfter)
	PtMax      *int
	Event      *RankEvent // Promotion, Demotion or NoEvent; nil when unknown
}

type replayHyp struct {
	rank, lo, hi int
	steps        []ReplayStep // games since the last point where only one hypothesis remained
}

// maxReplayHyps bounds the hypothesis set; ranges with the same rank are merged, so in practice
// it stays at 2–3.
const maxReplayHyps = 8

func ip(v int) *int { return &v }

func ep(e RankEvent) *RankEvent { return &e }

// ReplayTenhou replays a player's games for one mode.
func ReplayTenhou(games []ReplayGame) ([]ReplayStep, error) {
	out := make([]ReplayStep, len(games))
	var hyps []replayHyp
	pending := 0 // first game whose step is not yet written to out

	// flush writes the steps of games pending..end: exact if one hypothesis is left, otherwise the
	// fields all hypotheses agree on.
	flush := func(end int) {
		if pending >= end {
			return
		}
		for i := pending; i < end; i++ {
			var s *ReplayStep
			for _, h := range hyps {
				st := h.steps[i-pending]
				if s == nil {
					cp := st
					s = &cp
				} else {
					m := mergeStep(*s, st)
					s = &m
				}
			}
			if s != nil {
				out[i] = *s
			}
		}
		for k := range hyps {
			hyps[k].steps = nil
		}
		pending = end
	}

	for i, g := range games {
		if g.AnchorRank != nil {
			a := *g.AnchorRank
			r, err := tenhouRankAt(a)
			if err != nil {
				return nil, err
			}
			var keep []replayHyp
			for _, h := range hyps {
				if h.rank == a {
					keep = append(keep, h)
				}
			}
			if len(keep) == 0 {
				flush(i) // unreconcilable: keep what all old hypotheses agreed on, then restart
				hi := 0
				if r.PromotePt != nil {
					hi = *r.PromotePt - 1
				}
				keep = []replayHyp{{rank: a, lo: 0, hi: hi}}
			}
			hyps = keep
			if len(hyps) == 1 {
				flush(i)
			}
		}
		if len(hyps) == 0 {
			pending = i + 1
			continue
		}
		var next []replayHyp
		for _, h := range hyps {
			branches, err := applyHyp(h, g)
			if err != nil {
				return nil, err
			}
			next = append(next, branches...)
		}
		hyps = mergeHyps(next)
		if len(hyps) > maxReplayHyps { // shouldn't happen; give up until the next anchor
			flush(i + 1)
			hyps = nil
			continue
		}
		if len(hyps) == 1 {
			flush(i + 1)
		}
	}
	flush(len(games))
	return out, nil
}

// applyHyp applies one game to a hypothesis, splitting it where the pt range crosses a threshold.
func applyHyp(h replayHyp, g ReplayGame) ([]replayHyp, error) {
	r, err := tenhouRankAt(h.rank)
	if err != nil {
		return nil, err
	}
	with := func(rank, lo, hi, delta int, ev RankEvent) replayHyp {
		steps := make([]ReplayStep, len(h.steps), len(h.steps)+1)
		copy(steps, h.steps)
		steps = append(steps, ReplayStep{RankBefore: ip(h.rank), Delta: ip(delta), RankAfter: ip(rank),
			PtMin: ip(lo), PtMax: ip(hi), Event: ep(ev)})
		return replayHyp{rank: rank, lo: lo, hi: hi, steps: steps}
	}
	if r.PromotePt == nil { // 天鳳位: terminal
		return []replayHyp{with(h.rank, h.lo, h.hi, 0, NoEvent)}, nil
	}
	d, err := TenhouRankDelta(h.rank, TenhouGame{Players: g.Players, Level: g.Level, Length: g.Length, Placement: g.Placement})
	if err != nil {
		return nil, err
	}
	lo, hi, promote := h.lo+d, h.hi+d, *r.PromotePt
	var out []replayHyp
	if hi >= promote {
		start := tenhou.Ranks[h.rank+1].StartPt
		out = append(out, with(h.rank+1, start, start, d, Promotion))
	}
	stayLo, stayHi := max(lo, 0), min(hi, promote-1) // dan: below 0 demotes (branch below)
	if !r.CanDemote {
		stayHi = max(stayHi, 0) // kyu ranks floor at 0
	}
	if stayLo <= stayHi {
		out = append(out, with(h.rank, stayLo, stayHi, d, NoEvent))
	}
	if r.CanDemote && lo < 0 {
		start := tenhou.Ranks[h.rank-1].StartPt
		out = append(out, with(h.rank-1, start, start, d, Demotion))
	}
	return out, nil
}

// mergeHyps merges hypotheses with the same rank into one covering range.
func mergeHyps(hs []replayHyp) []replayHyp {
	var out []replayHyp
	for _, h := range hs {
		merged := false
		for k := range out {
			if out[k].rank == h.rank {
				out[k].lo, out[k].hi = min(out[k].lo, h.lo), max(out[k].hi, h.hi)
				for s := range out[k].steps {
					out[k].steps[s] = mergeStep(out[k].steps[s], h.steps[s])
				}
				merged = true
				break
			}
		}
		if !merged {
			out = append(out, h)
		}
	}
	return out
}

// mergeStep keeps the fields two possible histories agree on (pt as a covering range when the rank
// agrees).
func mergeStep(a, b ReplayStep) ReplayStep {
	eqInt := func(x, y *int) *int {
		if x != nil && y != nil && *x == *y {
			return x
		}
		return nil
	}
	s := ReplayStep{RankBefore: eqInt(a.RankBefore, b.RankBefore), Delta: eqInt(a.Delta, b.Delta),
		RankAfter: eqInt(a.RankAfter, b.RankAfter)}
	if s.RankAfter != nil && a.PtMin != nil && b.PtMin != nil && a.PtMax != nil && b.PtMax != nil {
		s.PtMin, s.PtMax = ip(min(*a.PtMin, *b.PtMin)), ip(max(*a.PtMax, *b.PtMax))
	}
	if a.Event != nil && b.Event != nil && *a.Event == *b.Event {
		s.Event = a.Event
	}
	return s
}
