# Related projects — capability analysis

Status: **Accepted (reference)** · Raw notes: [/docs/research/](../../research/)

## Amae Koromo (Mahjong Soul) — [research](../../research/amae-koromo.md)
- **Data:** crawler polls `fetchGameLiveList` per Gold+ room (~7 s), fetches `fetchGameRecord` after games end; stores in CouchDB. Frontend + crawler are MIT-licensed open source.
- **Player page:** basic (games, rank, avg rank, bust rate, placement rates), extended per-hand stats (win/deal-in/tsumo/dama/call/riichi rates, avg win/deal-in), riichi tab, luck tab, win/deal-in distribution, frequent opponents, percentile histograms, stable-rank estimate, last-30-games trend.
- **Access for us:** aggregate endpoints (`player_stats`, `player_extended_stats`) are open; `player_records` / `games` are **CAPTCHA-gated** (key from SAPikachu); **no full-log endpoint**.
- **Steal:** URL-encoded filter state, multi-mode select, formula tooltips, percentile histograms, "data since" annotations, stable-rank, delta & career leaderboards, AI-review deep link.
- **Gaps we fill:** no long-term rank chart, no per-player yaku breakdown, Gold+ only, 3p on a separate site, Chinese-first labels, no cross-platform view, weak mobile.

## Nodocchi (Tenhou) — [research](../../research/tenhou-and-nodocchi.md)
- **Data:** Tenhou `sc/raw` result files (all ranked lobbies, but log IDs only for 鳳凰) + 鳳凰 mjlogs for per-hand stats + live-spectate R ratings.
- **Player page:** estimated dan/pt & R with candlestick progression charts, stable dan, placement stats, avg pt, streaks, time-of-day heatmap, post-game-gap analysis, opponents, grade-change log, rich filters, CSV export.
- **Steal:** dan estimation by replaying games (constants documented), promotion/demotion log, candlestick rank chart, rich filters, CSV export.
- **Gaps we fill:** per-hand stats for non-鳳凰 players (via user-submitted logs), cross-platform view, accounts & ownership, modern UI, English.

## Converters & reviewers — [research](../../research/platforms-and-review-tools.md)
| Project | Use to us |
|---|---|
| tensoul (MIT) / Majsoul-to-NAGA (MIT) | Reference for Majsoul → tenhou.net/6 conversion |
| CityLogs `rc.honk.li` | RC → tenhou.net/6 API (post-v1) |
| cht33 RC parser, EpicOrange converter, InjusticeJudge | No licence → read for understanding, reimplement |
| Euophrys/tenhou-log, Benawii collector | Tenhou localStorage / Windows log list formats |
| NAGA | Paid, no API → generate tenhou.net/6 file + deep link |
| mjai.ekyu.moe (Mortal) | Free, accepts Tenhou/Majsoul URLs & tenhou.net/6 JSON → link out |
| Gokujan (ex-Bigcoach) | Paid since 2026-06, supports MS/Tenhou/RC → link out |
