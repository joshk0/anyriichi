# Tenhou ingestion

Status: **Proposed** · Research: [/docs/research/tenhou-and-nodocchi.md](../../research/tenhou-and-nodocchi.md)

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
