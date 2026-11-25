# Research: Amae Koromo (雀魂牌谱屋)

> Researched 2026-11-25. Sources: frontend `SAPikachu/amae-koromo` (MIT), crawler `SAPikachu/amae-koromo-scripts` (MIT), `amae-koromo-couchdb` design docs, a few GETs to the live API.
> **[V]** = verified in source or live response, **[I]** = inferred.

## 1. Player page stats

Component: `src/components/playerDetails/playerDetails.tsx`. Stat keys are Chinese strings returned verbatim by the API; labels in `src/locales/{en,ja}.json`. **[V]**

### Basic tab (基本)
- From `player_stats` (`PlayerMetadata`): `count` 记录场数 (recorded matches), `level` 记录等级/记录分数 (current rank / points), `avg_rank` 平均顺位, `negative_rate` 被飞率 (busting). Payload also has `rank_rates` (len 4 or 3), `rank_avg_score` (avg final score per placement), `max_level`, `played_modes`.
- From extended stats: 和牌率 win, 放铳率 deal-in, 自摸率 tsumo, 默听率 (dama), 流局率 exhaustive draw, 流听率 tenpai at draw, 副露率 call, 立直率 riichi, 和了巡数 avg win turn, 平均打点 avg win, 平均铳点 avg deal-in.
- **安定段位 (stable rank)** and **分数期望 (expected pt/game)** computed client-side from rank_rates + rank_avg_score + uma/penalty tables (`metadata.ts`), with games-to-rank-up/down estimate.

### Riichi tab (立直)
立直率, 立直后和牌率, 立直后放铳率 (A: incl. deal-in on riichi tile), 立直后非瞬间放铳率 (B: excl.), 立直收支/收入/支出, 先制率 first riichi, 追立率 chasing, 被追率 chased, 立直巡目 avg turn, 立直后流局率, 一发率, 振听立直率, 立直多面 multi-sided wait (since 2021-09-10), 立直好型2 good shape ≥6 tiles left (since 2021-11-07).

### More tab (更多)
最高等级/最高分数, 最大连庄 max repeats, 里宝率 uradora, 被炸率 + 平均被炸点数 (hit by dealer tsumo ≥ mangan), 放铳时立直率/副露率, 副露后放铳率/和牌率/流局率, 打点效率 (win% × avg win), 铳点损失 (deal-in% × avg deal-in), 净打点效率 (difference), 局收支 G/L per round, 总计局数 total rounds.

### Win/deal-in distribution (和铳分布)
Pies: wins by riichi/open/dama; own state at deal-in; dealt into riichi/open/dama.

### Luck (血统)
役满 count, 累计役满, 最大累计番数, 流满, W立直, 平均起手向听 haipai shanten (overall/dealer/non-dealer, since 2022-06-27).

### Other
- 最近大铳: most recent ≥ mangan deal-in with yaku + replay link.
- 最常同桌 frequent opponents: from last 100 games (count, H2H %, avg placement, both players' distributions).
- **Percentile histogram per stat** on hover (`global_histogram`), when one mode selected & > 100 rounds.
- No per-player yaku breakdown (site-wide `fan_stats` only). **[V]**

## 2. Filters & modes **[V]**

| | Gold 金 | Jade 玉 | Throne 王座 |
|---|---|---|---|
| 4p South | 9 | 12 | 16 |
| 4p East | 8 | 11 | 15 |
| 3p South | 22 | 24 | 26 |
| 3p East | 21 | 23 | 25 |

- Gold+ only (no Bronze/Silver/friendly/tournament). 3p is a separate site `ikeda.sapk.ch` (`api/v2/pl3/`), 4p `api/v2/pl4/`. Data since 2019-08-23 (4p), 2019-11-29 (3p).
- Level IDs: `[1=4p|2=3p]` + major + minor, e.g. 10502 = 4p Saint 2. Majors 初 Novice, 士 Adept, 杰 Expert, 豪 Master, 圣 Saint, 魂 Celestial; major 7 = new Celestial (Konten) with its own point table.
- Max points per level, 4th-place penalties, uma (`MODE_DELTA`) hard-coded in `level.ts` / `metadata.ts`. Delta: `ceil((score-25000)/1000 + [15,5,-5,-15][rank]) + MODE_DELTA[mode][rank]` minus level penalty on 4th.
- Server zone from account id: `id >> 23` → 0–6 CN, 7–12 JP, 13–15 International.
- URL state: `/player/:id/:modes(16.12…)/:-search?/:startDate?/:endDate?` + `?rank=N&kontenOnly=1&limit=N`. Multi-mode joined with `.`; mode choice remembered per player.
- Date presets: all, custom range, last 4/13/26/52 weeks, this/last month/year, last N games (100…5000), new/old Throne (split 2021-08-26).
- Extra filters: placement, opponent name, Konten-only. Enabling any hides stable-rank/best-rank.
- Charts: "Trends" line chart of last 30 games (placement + cumulative rank-pt delta), dots coloured by placement, click → replay. Placement pie. **No long-term rank-point history.**

## 3. Other pages
- Home: player prefix search (`search_player`), browse finished games by date/mode.
- Player games table (mode, placement, 4 players w/ rank badge & score, datetime), infinite scroll ×100.
- `/highlight`: recent yakuman.
- `/ranking`: delta ranking (1d/3d/1w/4w gainers/losers/most games); career rankings (min 300 games) across ~20 metrics.
- `/statistics`: placement by seat, stats by rank (overall/yearly/snapshot), site-wide yaku frequency, player counts per level & zone.
- Game actions: replay, copy link, player details, AI review deep link (mjai.ekyu.moe / Mortal).
- Starred players (localStorage). UI in zh/ja/en/ko. Contest variant subdomains (`api/contest/{id}/`).
- No live-games page.

## 4. API & data pipeline

- API mirrors: `https://{1,2,4,5}.data.amae-koromo.com/api/v2/pl4/` (client fails over). Responses may be `{maintenance}` or `{result_key}` (async job polled at `result/{key}`).
- Endpoints (ms epoch timestamps, modes dot-joined):
  - `player_stats/{id}/{start}/{end}?mode=16.12&tag={hour}` — **open (200)**
  - `player_extended_stats/{id}/{start}/{end}?mode=` — **open (200)**
  - POST `player_stats/{id}` / `player_extended_stats/{id}` body `{keys:[startTimes], modes}`
  - `player_records/{id}/{cursorEnd}/{start}?limit=100&mode=&descending=true` — **CAPTCHA-gated (429)**
  - `games/{cursor}/{dayStart}?limit=100…` — **CAPTCHA-gated**
  - `search_player/{prefix}?limit=20&tag=all`, `games_by_id/{ids}`, `recent_highlight_games?mode=`, `player_delta_ranking/{1d|3d|1w|4w}`, `career_ranking/{type}[_{min}]?mode=`, `global_statistics_2`, `global_statistics_year`, `global_statistics_snapshot/{date}`, `global_histogram`, `fan_stats`, `rank_rate_by_seat`, `level_statistics`, `view_game/{zone}/{mode}/{_id}`.
- `GameRecord`: `{_id, modeId, uuid, startTime, endTime, players:[{accountId, nickname, level, score, gradingScore}]}` (seconds).
- **Gate:** 429 `[x-cap-token-required] CAPTCHA required, please contact SAPikachu to obtain a key if you are crawling the data`. Frontend solves Cap.js PoW (WASM) → `Authorization: Bearer`.
- **No public endpoint serves full game logs** (records are stored server-side only). **[V]**

### Crawler (`amae-koromo-scripts/logGames.js`, `majsoul.js`) **[V]**
- Logs into Majsoul via websocket protobuf (`.lq.Route.requestConnection`, `Lobby.oauth`/`login`, `heatbeat`); proto from `fetchLatestDataDefinition`.
- Polls `.lq.Lobby.fetchGameLiveList({filter_id: 200+mode})` for each mode ~every 7 s.
- Records each live game UUID as pending; after it ends, `.lq.Lobby.fetchGameRecord({game_uuid, client_version_string})` (head + protobuf data, maybe via `data_url`). Retries; drops after 48 h.
- Stores `records/{date}/{mode}/{modeId}/{uuid}.json` + `.recordData` (optional GCS).
- `gameAnalyzer.ts` replays `RecordNewRound`, `DealTile`, `DiscardTile`, `ChiPengGang`, `Hule`, `NoTile`, `LiuJu` with shanten/agari libs to compute round stats. **Good reference for our stat definitions.**
- CouchDB/PouchDB per-mode DBs, map/reduce views; SQLite cache; Express API (`extApi.js`, drifted from prod).
- [I] International (Yostar) server: record fetching disabled there; records presumably fetched via CN/JP.

## 5. UX worth copying / gaps

**Copy:** shareable URL state; multi-mode select remembered per player; tabbed stats w/ formula tooltips; percentile histograms; "data since" annotations; stable-rank & expected-pt; trend dots → replay; consistent placement colours; computed rank deltas; leaderboards; AI review deep link.

**Gaps (our opportunities):** no long-term rank history; no per-player yaku breakdown; no live view / notifications; Gold+ only; 3p on separate site; Chinese-first keys and awkward English; no player-vs-player compare; nothing beyond riichi metrics (efficiency delegated to Mortal); no bulk API; basic charts, weak mobile.
