# Tenhou ingestion

Status: **Accepted — prototype implemented** (`worker/internal/{tenhou,ingest}`, migration `20261126000000_tenhou_ingest.sql`); dan/R replay still Proposed · Research: [/docs/research/tenhou-and-nodocchi.md](../../research/tenhou-and-nodocchi.md)

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

## Dan / R estimation
Replay each player's ranked games in order with Nodocchi's constants (statistics-model §2). Anchor with `<UN dan= rate=>` from any full log. Track promotions/demotions → `rank_event`.

## Compliance
≤ 1 poll / 20 min, gz only, URLs from `/sc/raw/dat/`, single session, no redistribution of files or index, no public log links (D-007).

## Prototype implementation (2026-11-26)
- **Poll** (`tenhou-raw-poll`, scheduler): last poll time persisted in `private.worker_state`, so restarts/deploys never poll more often than every 20 min. Ingests hourly `scb`/`scc` files newer than `TENHOU_BACKFILL` (default 48 h, keeps the free-tier dev DB small) whose size grew, oldest first, one file at a time; one transaction + one batched round trip per file via `private.upsert_tenhou_summary()`.
- **Identity:** `games.match_key` = md5(start hour JST | rule code | name|pt×10 in placement order). The same 鳳凰 game in `scb`, `scc` and its full log collapses into one row (`sources` = {log, scb, scc}). Start minute/duration come from summaries.
- **Full logs** (`tenhou.log` queue kind, concurrency 1): from `scc` (`TENHOU_FETCH_HOUOU_LOGS`, default on) and submissions (`worker enqueue <id|url>` for now; the web submit page will call `queue`/`private.tenhou_logs` the same way). Rejected: invalid id, non-ranked lobby or shuugi, logs before 2017-10-25 JST (D-046), GO lobby ≠ 0. Raw XML stored gzipped in `private.tenhou_logs.raw_gz` (prototype; Supabase Storage later, D-039). Parsed into `private.hands` / `private.hand_player_stats` (win/deal-in/riichi/call/tenpai/han/fu/value/yaku/score delta).
- **HTTP:** one shared client, serialised, ≥ 2 s between requests, generic User-Agent, URLs built from `/sc/raw/dat/` and `/0/log/`.
- **Rank pt:** stored when computable — non-last placements always; last place only from full logs (rank from `<UN dan>`). 天鳳位 = 0. Promotion/demotion replay is the next step.
- **Privacy:** log ids, raw logs and per-hand rows live in the `private` schema (not exposed by the Data API, D-007/D-029); `players`, `rulesets`, `games`, `game_participants` are public-read via RLS.
- **Test fixtures are synthetic** — the repo is public and Tenhou forbids redistributing result files and logs.
