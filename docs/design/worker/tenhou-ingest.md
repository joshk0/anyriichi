# Tenhou ingestion

Status: **Accepted — prototype implemented** (`worker/internal/{tenhou,ingest}`, migration `20261126000000_tenhou_ingest.sql`); dan/pt replay implemented (D-050); R replay later (#59) · Research: [/docs/research/tenhou-and-nodocchi.md](../../research/tenhou-and-nodocchi.md)

## Public results (all games, D-001)
1. Every 20 min: GET `https://tenhou.net/sc/raw/list.cgi` (JSONP) → `[{file,size}]`.
2. For each file in scope (`scb*`, `scc*` only — D-033) whose `size` > stored size: download `https://tenhou.net/sc/raw/dat/{file}` (gz only), follow redirects, one connection at a time.
3. Parse lines → upsert `games` (+ `rulesets` from rule string) and `game_participants` (placement order, ±pt). `scc` lines include the log id.
4. Store `(file, size, etag/last-modified, processed_at)` in `tenhou_raw_files`.
5. Backfill: `list.cgi?old` (~1 year of daily files) once at launch, slowly. Older history: Q-031.

## Full logs
- Sources: user submissions, extension, premium import, and `scc` (鳳凰) log ids.
- Fetch `https://tenhou.net/0/log/?{logid}` (mjlog XML; CORS-open, no auth), store raw privately, parse → hands / hand_player_stats, convert to tenhou.net/6 JSON.
- Rate limit ≈ 1 request / 2 s; retry with backoff; never parallel.
- Match to existing `scb` summary game (see game-record-model).
- Validate ranked (D-012): rule type & lobby `0000`.

## Dan / pt replay (D-050, `worker/internal/rules/replay.go`)
Tenhou never publishes rank points, and full logs only show each player's **rank at game start** (`<UN dan=…>`).
The replay walks each player's ranked games in order, separately for 3p and 4p, keeping a small set of
hypotheses `(rank, pt range)`:
- **Unknown until the first anchor** (a game with a full log). We don't assume a starting rank, because our data window starts mid-career.
- **Anchor:** hypotheses with another rank are dropped. If none survive (first anchor, or games we can't
  see), the player restarts at that rank with pt anywhere in `[0, promotePt)`.
- **Each game** shifts every range by that game's rank-pt change (rules §2). A range that crosses a promotion
  or demotion threshold splits into "stayed" and "promoted/demoted" hypotheses. A later anchor shows which one happened, and the
  earlier game's event is filled in after the fact. Promotion and demotion reset pt to the new rank's start, so from then on pt is **exact**.
- Kyu ranks floor at 0. 天鳳位 is terminal.

Results go to `game_participants`: `est_rank_after`, `est_pt_min`/`est_pt_max` (the range after the game),
`rank_event` (`promotion`/`demotion`), and `rank_pt_delta` once the rank before the game is certain (this also fills last-place
deltas for games known only from summaries). Ingest queues affected players in `private.rank_replay_queue`, and the
`tenhou-rank-replay` scheduler job (every minute) replays them in batches of 500. `worker replay` re-runs everyone.
**First production run (2026-11-25, ~2 days of data):** 22k players in under 20 s. Rank is known for 94 % of 鳳凰 rows.
Exact pt comes only after a promotion or demotion is observed, so it is rare until more history accumulates.

Tenhou R replay (D-026) comes later (#59). Seed it from `<UN rate>` the same way.

## Compliance
≤ 1 poll / 20 min, gz only, URLs from `/sc/raw/dat/`, single session, no redistribution of files or index, no public log links (D-007).

## Prototype implementation (2026-11-26)
- **Poll** (`tenhou-raw-poll`, scheduler): last poll time persisted in `private.worker_state`, so restarts/deploys never poll more often than every 20 min. Ingests hourly `scb`/`scc` files newer than `TENHOU_BACKFILL` (default 48 h, keeps the free-tier dev DB small) whose size grew, oldest first, one file at a time; one transaction + one batched round trip per file via `private.upsert_tenhou_summary()`.
- **Identity:** `games.match_key` = md5(start hour JST | rule code | name|pt×10 in placement order). The same 鳳凰 game in `scb`, `scc` and its full log collapses into one row (`sources` = {log, scb, scc}). Start minute/duration come from summaries.
- **Full logs** (`tenhou.log` queue kind, concurrency 1): from `scc` (`TENHOU_FETCH_HOUOU_LOGS`, default on) and submissions (`worker enqueue <id|url>` for now; the web submit page will call `queue`/`private.tenhou_logs` the same way). Rejected: invalid id, non-ranked lobby or shuugi, logs before 2017-10-25 JST (D-046), GO lobby ≠ 0. Raw XML stored gzipped in `private.tenhou_logs.raw_gz` (prototype; Supabase Storage later, D-039). Parsed into `private.hands` / `private.hand_player_stats` (win/deal-in/riichi/call/tenpai/han/fu/value/yaku/score delta).
- **HTTP:** one shared client, serialised, ≥ 2 s between requests, generic User-Agent, URLs built from `/sc/raw/dat/` and `/0/log/`.
- **Rank pt:** stored when computable — non-last placements always; last place from full logs (rank from `<UN dan>`) or once the replay knows the rank. 天鳳位 = 0.
- **Privacy:** log ids, raw logs and per-hand rows live in the `private` schema (not exposed by the Data API, D-007/D-029); `players`, `rulesets`, `games`, `game_participants` are public-read via RLS.
- **Test fixtures are synthetic** — the repo is public and Tenhou forbids redistributing result files and logs.
