# Statistics model

Status: **Proposed** (D-016 accepted; tables to verify Q-042) · Related: [game record model](game-record-model.md)

## 1. Normalized result (±pt, D-016)
For every participant: `result_pt = (final_score − return_points) / 1000 + uma[placement] (+ oka to 1st)`,
where `oka = (return_points − start_points) × players / 1000`. Tenhou already publishes this value (use as-is).
Ties: split uma evenly (Q-043). Tables per ruleset live in `/packages/shared/rules/*.json`.

| Ruleset | start / return | uma | Source |
|---|---|---|---|
| Tenhou 4p | 25k / 30k | +20/+10/−10/−20 | Nodocchi, Tenhou |
| Tenhou 3p | 35k / 40k | +20/0/−20 | |
| Majsoul 4p | 25k / 25k? | +15/+5/−5/−15 (score screen) | to verify (Q-042) |
| Majsoul 3p | 35k / 35k? | +15/0/−15 | to verify |
| Riichi City 4p (later) | 25k / 30k | +30/+10/−10/−30 | |

## 2. Rank points per game
- **Tenhou** (Nodocchi constants): gains 4p 1st/2nd (3p 1st) — 般 +20/+10 (+30), 上 +40/+10 (+50), 特 +50/+20 (+70), 鳳 +60/+30 (+90); 南 ×1.5; last-place loss by dan (初段 −30 … 十段 −120, kyu 0/−10/−20); thresholds per dan; historical variants before 2017-10-24. Dan is **estimated** by replaying the player's ranked games from their first; anchor when a known dan is observed (Tenhou mjlog `<UN dan=...>`).
- **Majsoul**: `ceil((score − 25000)/1000 + [15,5,−5,−15][place]) + MODE_DELTA[mode][place]`, minus level penalty on 4th (AK `level.ts`); Majsoul records also carry `gradingScore` → prefer the source value.
- **Tenhou R (D-026, later):** start 1500; ΔR = (placement base + (table avg R − R)/40) × max(1 − 0.002·games, 0.2); base 4p +30/+10/−10/−30, 3p +30/0/−30 (verify). Seed/re-anchor from `<UN rate>` in full logs; store `rating_before`/`rating_delta` on `game_participants`. Not shown for Mahjong Soul.
- `rank_event` set when a game crosses a promotion/demotion threshold → highlighted in UI.

## 3. Placement stats (summary tier — all games)
games, 1st/2nd/3rd/4th %, avg placement, top-2 %, bust % (Majsoul score < 0; Tenhou inferred from log), avg ±pt, total ±pt, stable rank estimate (expected rank from placement rates & rank tables), current/peak rank.

## 4. Per-hand stats (full-log tier)
Defined as ratios over `hand_player_stats`. Initial set (AK parity): win rate, deal-in rate, tsumo rate, dama rate, call rate, riichi rate, draw rate, tenpai-at-draw rate, avg win turn, avg win value, avg deal-in value, riichi win/deal-in (A & B), first/chasing/chased riichi, avg riichi turn, ippatsu, uradora, furiten riichi, multi-sided wait, win/deal-in distribution by state, luck (yakuman, double riichi, haipai shanten), win efficiency, deal-in loss, net efficiency, G/L per round.
Each stat records the **coverage** (how many of the player's games have full logs) — shown in UI.

## 5. Aggregation strategy
- Source of truth: row-level tables. Aggregation by SQL with filters (platform, ruleset attributes, date range, last-N).
- Cache: `stats_cache(player_id, filter_hash, computed_at, payload jsonb)`, invalidated per player on new games.
- Population percentiles (AK histogram): nightly job computes per-ruleset, per-rank distributions into `stat_distributions`.
