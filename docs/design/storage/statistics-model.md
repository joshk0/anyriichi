# Statistics model

Status: **Accepted** for §1–2 (D-016, D-045–D-047; implemented in `packages/shared/src/rules.ts` + `worker/internal/rules`, data in `packages/shared/rules/*.json`) · rest **Proposed** · Related: [game record model](game-record-model.md)

## 1. Normalized result (±pt, D-016)
For every participant: `result_pt = (final_score − return_points) / 1000 + uma[placement] (+ oka to 1st)`,
where `oka = (return_points − start_points) × players / 1000`. Tenhou already publishes this value (use as-is).
Ties: both platforms break equal scores by seat order (closer to the initial dealer ranks higher), so we use the platform's placement as-is — no uma splitting (resolves Q-043). Tables live in `packages/shared/rules/*.json`.

**No cross-platform ±pt (D-047):** Tenhou and Mahjong Soul ±pt are on different scales, so ±pt is only ever shown/aggregated per platform. Combined ("All platforms") views show placement stats only.

| Ruleset | start / return | uma | Source |
|---|---|---|---|
| Tenhou 4p | 25k / 30k | +20/+10/−10/−20 | Nodocchi, Tenhou |
| Tenhou 3p | 35k / 40k | +20/0/−20 | tenhou.net/man |
| Majsoul 4p | 25k / 25k | +15/+5/−5/−15 | Amae Koromo `metadata.ts` (Q-042 resolved) |
| Majsoul 3p | 35k / 35k | +15/0/−15 | Amae Koromo `metadata.ts` |
| Riichi City 4p (later) | 25k / 30k | +30/+10/−10/−30 | |

## 2. Rank points per game
- **Tenhou** (source: tenhou.net/man 段級位制; post-2017-10-24 rules only, D-046): 4p non-last gains by lobby, East / South — 一般 +20/+10/0 · +30/+15/0, 上級 +40/+10/0 · +60/+15/0, 特上 +50/+20/0 · +75/+30/0, 鳳凰 +60/+30/0 · +90/+45/0; 3p 1st — 一般 +30·+45, 上級 +50·+75, 特上 +70·+105, 鳳凰 +90·+135 (2nd 0). Last place by rank, same for 3p/4p: 2級 −10/−15, 1級 −20/−30, 初段 −30/−45 … 十段 −120/−180 (−10/−15 per dan); 3級 and below 0. Promotion at 20/20/20/20/40/60/80/100/100/100 (新人…1級) and 400…4000 (初段…十段, +400 each); new rank starts at 0 (kyu) or 200…2000 (dan), surplus discarded. Dan ranks demote below 0 to the previous rank's start; kyu ranks floor at 0. 十段 → 天鳳位 is terminal. 3p and 4p ranks are tracked separately. Dan is **estimated** by replaying ranked games; anchor when a known dan is observed (Tenhou mjlog `<UN dan=...>`).
- **Mahjong Soul** (source: Amae Koromo `level.ts`/`metadata.ts`, MIT): `ceil((score − base)/1000 + uma[place]) + placementBonus[mode][place]`, minus the level penalty on last place (separate tables for 4p/3p × South/East). Gold/Jade/Throne, plus 4p Bronze (+10/+5 E, +20/+10 S) and Silver (+20/+10 E, +40/+20 S) from riichi.wiki/npmahjong (mode ids 2/3/5/6 inferred). 3p Bronze/Silver have no verified table: their rank delta comes only from the record's grading score (D-048). Level max points 20/80/200 (Novice), 600/800/1000, 1200/1400/2000, 2800/3200/3600, 4000/6000/9000 (Saint); promotion/demotion resets to half the new level's max (Novice 0); Novice and Adept 1 never demote. Konten (major 7, levels 1–20): max 20.0 (2000), start 10.0, fixed Throne deltas (4p S +50/+20/−20/−50, E +30/+10/−10/−30; 3p S ±50, E ±30); Konten 1 below 0 → Saint 3. Records also carry `gradingScore` → prefer the source value when present.
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
