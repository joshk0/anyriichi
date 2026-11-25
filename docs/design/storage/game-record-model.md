# Game record model

Status: **Proposed** · Related: [statistics model](statistics-model.md), [log conversion](../worker/log-conversion.md)

Three tiers of data per game, because sources differ in richness:

| Tier | Source examples | Contains |
|---|---|---|
| **Summary** | Tenhou `scb`/`scc` lines, AK `player_records` | players, placement, final pt / score, room, start time, duration |
| **Full log** | Tenhou mjlog, Majsoul `fetchGameRecord` | every draw/discard/call/win → per-hand stats |
| **Derived** | our parser | per-hand rows, per-game per-player stat counters |

```sql
rulesets(                    -- one row per distinct room/rule combination
  id serial PK, platform text,
  code text,                 -- tenhou rule type hex '00a9' | majsoul mode id '12'
  players smallint,          -- 3 | 4
  length text,               -- east | south
  lobby text,                -- ippan/joukyuu/tokujou/houou | bronze/silver/gold/jade/throne
  ranked boolean, aka boolean, kuitan boolean, speed text,
  start_points int, return_points int, uma int[],   -- for normalization (D-016)
  unique(platform, code))

games(
  id bigserial PK, platform text, external_id text,   -- tenhou log id | majsoul uuid | synthetic for scb rows
  has_external_id boolean,          -- scb rows have no log id (they're matched later)
  ruleset_id int references rulesets,
  lobby_ref text null,              -- tenhou L/C lobby number, majsoul contest id
  started_at timestamptz, duration_min smallint null,
  source text,                      -- scb | scc | sca | scf | ak | fetch | submission
  full_log_status text,             -- none | queued | stored | failed
  raw_log_path text null,           -- private object-storage key
  unique(platform, external_id))

game_participants(
  game_id bigint, seat smallint null, placement smallint,
  player_id bigint references players,
  final_score int null,             -- raw points (e.g. 48300); null when only pt known (tenhou scb)
  result_pt numeric(6,1),           -- normalized tenhou-style ±pt (D-016)
  rank_before text null, rank_pt_before int null,
  rank_pt_delta int null,           -- D-016
  rank_event text null,             -- promoted | demoted | null
  rating_before numeric null,       -- tenhou R if known
  rating_delta numeric null,        -- tenhou ΔR (D-026)
  primary key (game_id, placement))

hands(                              -- one row per kyoku per game (full logs only)
  game_id, hand_index, round_wind, dealer_seat, honba, riichi_sticks,
  outcome text,                     -- ron | tsumo | ryuukyoku | abortive
  primary key (game_id, hand_index))

hand_player_stats(                  -- one row per hand per seat: the atoms of every per-hand stat
  game_id, hand_index, seat, player_id,
  won bool, won_by text, dealt_in bool, dealt_into_seat smallint,
  riichi bool, riichi_turn smallint, riichi_first bool, riichi_chasing bool, riichi_wait_count smallint,
  furiten_riichi bool, ippatsu bool, uradora smallint,
  called bool, call_count smallint, dama_win bool,
  tenpai_at_draw bool, win_turn smallint, win_value int, deal_in_value int, han smallint, fu smallint,
  yaku smallint[], start_shanten smallint, score_delta int,
  primary key (game_id, hand_index, seat))
```

## Matching summary rows to full logs (Tenhou)
`scb` rows have no log id. When a full log arrives (submission / scc), match it to the existing summary game by
(start hour from log id, rule type, the 4 player names) and attach `external_id`; otherwise insert new.

## Idempotency
Every ingest is an upsert keyed on `(platform, external_id)` or the match key above; re-running a file is safe.
