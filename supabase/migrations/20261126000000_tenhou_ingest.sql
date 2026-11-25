-- Tenhou ingestion prototype (docs/design/worker/tenhou-ingest.md, docs/design/storage/game-record-model.md).
--
-- Public schema: data that is public by nature (derived from Tenhou's public results, D-030).
--   Readable by everyone through the Data API; written only by the worker (table owner).
-- Private schema: log ids, raw logs and per-hand detail. Log ids are replay links and follow
--   per-player visibility (D-007, D-029); never expose them through the Data API.

create schema if not exists private;

-- --------------------------------------------------------------------------- public

create table public.players (
  id            bigint generated always as identity primary key,
  platform      text not null check (platform in ('tenhou', 'majsoul', 'riichicity')),
  platform_key  text not null,          -- tenhou: player name
  display_name  text not null,
  first_seen_at timestamptz not null,
  last_seen_at  timestamptz not null,
  unique (platform, platform_key)
);

create table public.rulesets (
  id       integer generated always as identity primary key,
  platform text not null,
  code     text not null,                -- tenhou: 4-hex rule type, e.g. '00a9'
  players  smallint not null check (players in (3, 4)),
  length   text not null check (length in ('east', 'south')),
  lobby    text not null,                -- tenhou: ippan | joukyuu | tokujou | houou
  ranked   boolean not null,
  aka      boolean not null,
  kuitan   boolean not null,
  speed    text not null default 'normal',
  unique (platform, code)
);

create table public.games (
  id            bigint generated always as identity primary key,
  platform      text not null,
  ruleset_id    integer not null references public.rulesets,
  -- Identity across sources without the log id: md5 of start hour, rule code and
  -- players with ±pt in placement order (docs/design/storage/game-record-model.md).
  match_key     text not null,
  started_at    timestamptz not null,
  duration_min  smallint,
  sources       text[] not null default '{}',   -- scb | scc | log
  has_full_log  boolean not null default false,
  created_at    timestamptz not null default now(),
  updated_at    timestamptz not null default now(),
  unique (platform, match_key)
);
create index games_started_at_idx on public.games (started_at desc);

create table public.game_participants (
  game_id       bigint not null references public.games on delete cascade,
  placement     smallint not null check (placement between 1 and 4),
  seat          smallint check (seat between 0 and 3),   -- known from full logs only
  player_id     bigint references public.players,        -- null for guests (NoName)
  name          text not null,
  final_score   integer,                                 -- known from full logs only
  result_pt     numeric(6, 1) not null,                  -- Tenhou-published ±pt (D-016)
  rank_before   integer,      -- tenhou rank index (rules: 0 = 新人 … 20 = 天鳳位), from full logs
  rating_before numeric(7, 2),                           -- tenhou R at game start, from full logs
  rank_pt_delta integer,      -- null when it depends on an unknown rank (last place, summary only)
  primary key (game_id, placement)
);
create index game_participants_player_idx on public.game_participants (player_id, game_id);

alter table public.players enable row level security;
alter table public.rulesets enable row level security;
alter table public.games enable row level security;
alter table public.game_participants enable row level security;
create policy "public read" on public.players for select using (true);
create policy "public read" on public.rulesets for select using (true);
create policy "public read" on public.games for select using (true);
create policy "public read" on public.game_participants for select using (true);

-- --------------------------------------------------------------------------- private

create table private.worker_state (
  key        text primary key,
  value      jsonb not null,
  updated_at timestamptz not null default now()
);

create table private.tenhou_raw_files (
  file          text primary key,       -- e.g. scb2026112603.log.gz
  size          bigint not null,
  lines         integer not null,
  processed_at  timestamptz not null default now()
);

create table private.tenhou_logs (
  log_id        text primary key,
  game_id       bigint references public.games on delete set null,
  status        text not null default 'queued' check (status in ('queued', 'stored', 'failed', 'rejected')),
  reject_reason text,
  source        text not null,          -- scc | submission | extension | import
  raw_gz        bytea,                  -- gzip of the mjlog XML; never served publicly
  fetched_at    timestamptz,
  created_at    timestamptz not null default now(),
  updated_at    timestamptz not null default now()
);
create index tenhou_logs_game_idx on private.tenhou_logs (game_id);

create table private.hands (
  game_id        bigint not null references public.games on delete cascade,
  hand_index     smallint not null,
  round          smallint not null,      -- 0 = East 1, 4 = South 1, …
  honba          smallint not null,
  riichi_sticks  smallint not null,
  dealer_seat    smallint not null,
  outcome        text not null check (outcome in ('ron', 'tsumo', 'draw', 'abortive')),
  primary key (game_id, hand_index)
);

create table private.hand_player_stats (
  game_id       bigint not null,
  hand_index    smallint not null,
  seat          smallint not null,
  player_id     bigint references public.players,
  won           boolean not null default false,
  won_by        text check (won_by in ('ron', 'tsumo')),
  dealt_in      boolean not null default false,
  riichi        boolean not null default false,
  called        boolean not null default false,
  tenpai_at_draw boolean,
  han           smallint,
  fu            smallint,
  win_value     integer,
  deal_in_value integer,
  yaku          smallint[],
  score_delta   integer not null default 0,
  primary key (game_id, hand_index, seat),
  foreign key (game_id, hand_index) references private.hands on delete cascade
);
create index hand_player_stats_player_idx on private.hand_player_stats (player_id);

-- One round trip per summary line: upserts ruleset, players, game and participants.
-- p_names / p_pt_tenths are in placement order. Returns the game id.
create function private.upsert_tenhou_summary(
  p_match_key   text,
  p_started_at  timestamptz,
  p_duration    smallint,
  p_source      text,
  p_code        text,
  p_players     smallint,
  p_length      text,
  p_lobby       text,
  p_aka         boolean,
  p_kuitan      boolean,
  p_speed       text,
  p_names       text[],
  p_pt_tenths   integer[],
  p_rank_deltas integer[]     -- -99999 = unknown (null)
) returns bigint
language plpgsql
set search_path = ''
as $$
declare
  v_ruleset integer;
  v_game    bigint;
  v_player  bigint;
  i         integer;
begin
  insert into public.rulesets (platform, code, players, length, lobby, ranked, aka, kuitan, speed)
  values ('tenhou', p_code, p_players, p_length, p_lobby, true, p_aka, p_kuitan, p_speed)
  on conflict (platform, code) do update set code = excluded.code
  returning id into v_ruleset;

  insert into public.games (platform, ruleset_id, match_key, started_at, duration_min, sources)
  values ('tenhou', v_ruleset, p_match_key, p_started_at, p_duration, array[p_source])
  on conflict (platform, match_key) do update set
    duration_min = coalesce(public.games.duration_min, excluded.duration_min),
    started_at   = excluded.started_at,   -- summaries carry the start minute; full logs only the hour
    sources      = (select array(select distinct unnest(public.games.sources || excluded.sources) order by 1)),
    updated_at   = now()
  returning id into v_game;

  for i in 1 .. array_length(p_names, 1) loop
    v_player := null;
    if p_names[i] <> 'NoName' then
      insert into public.players (platform, platform_key, display_name, first_seen_at, last_seen_at)
      values ('tenhou', p_names[i], p_names[i], p_started_at, p_started_at)
      on conflict (platform, platform_key) do update set
        first_seen_at = least(public.players.first_seen_at, excluded.first_seen_at),
        last_seen_at  = greatest(public.players.last_seen_at, excluded.last_seen_at)
      returning id into v_player;
    end if;

    insert into public.game_participants (game_id, placement, player_id, name, result_pt, rank_pt_delta)
    values (v_game, i, v_player, p_names[i], p_pt_tenths[i] / 10.0, nullif(p_rank_deltas[i], -99999))
    on conflict (game_id, placement) do update set
      player_id     = excluded.player_id,
      name          = excluded.name,
      result_pt     = excluded.result_pt,
      rank_pt_delta = coalesce(public.game_participants.rank_pt_delta, excluded.rank_pt_delta);
  end loop;

  return v_game;
end;
$$;

-- Full log: like upsert_tenhou_summary plus seats, final scores and start-of-game rank/R.
-- Arrays are in placement order. Keeps a summary's start minute and duration if present.
create function private.upsert_tenhou_full_log(
  p_match_key   text,
  p_started_at  timestamptz,
  p_code        text,
  p_players     smallint,
  p_length      text,
  p_lobby       text,
  p_aka         boolean,
  p_kuitan      boolean,
  p_speed       text,
  p_names       text[],
  p_pt_tenths   integer[],
  p_rank_deltas integer[],
  p_seats       integer[],
  p_scores      integer[],
  p_dans        integer[],
  p_rates       numeric[]
) returns bigint
language plpgsql
set search_path = ''
as $$
declare
  v_ruleset integer;
  v_game    bigint;
  v_player  bigint;
  i         integer;
begin
  insert into public.rulesets (platform, code, players, length, lobby, ranked, aka, kuitan, speed)
  values ('tenhou', p_code, p_players, p_length, p_lobby, true, p_aka, p_kuitan, p_speed)
  on conflict (platform, code) do update set code = excluded.code
  returning id into v_ruleset;

  insert into public.games (platform, ruleset_id, match_key, started_at, sources, has_full_log)
  values ('tenhou', v_ruleset, p_match_key, p_started_at, array['log'], true)
  on conflict (platform, match_key) do update set
    has_full_log = true,
    sources      = (select array(select distinct unnest(public.games.sources || excluded.sources) order by 1)),
    updated_at   = now()
  returning id into v_game;

  for i in 1 .. array_length(p_names, 1) loop
    v_player := null;
    if p_names[i] <> 'NoName' then
      insert into public.players (platform, platform_key, display_name, first_seen_at, last_seen_at)
      values ('tenhou', p_names[i], p_names[i], p_started_at, p_started_at)
      on conflict (platform, platform_key) do update set
        first_seen_at = least(public.players.first_seen_at, excluded.first_seen_at),
        last_seen_at  = greatest(public.players.last_seen_at, excluded.last_seen_at)
      returning id into v_player;
    end if;

    insert into public.game_participants
      (game_id, placement, seat, player_id, name, final_score, result_pt, rank_before, rating_before, rank_pt_delta)
    values (v_game, i, p_seats[i], v_player, p_names[i], p_scores[i], p_pt_tenths[i] / 10.0,
            p_dans[i], p_rates[i], nullif(p_rank_deltas[i], -99999))
    on conflict (game_id, placement) do update set
      seat          = excluded.seat,
      player_id     = excluded.player_id,
      name          = excluded.name,
      final_score   = excluded.final_score,
      result_pt     = excluded.result_pt,
      rank_before   = excluded.rank_before,
      rating_before = excluded.rating_before,
      rank_pt_delta = excluded.rank_pt_delta;
  end loop;

  return v_game;
end;
$$;

-- Lock the private schema down (the worker connects as the owner).
revoke all on schema private from public;
revoke all on all tables in schema private from public;
revoke all on all functions in schema private from public;

do $$
declare r text;
begin
  foreach r in array array['anon', 'authenticated'] loop
    if exists (select 1 from pg_roles where rolname = r) then
      execute format('revoke all on schema private from %I', r);
      execute format('revoke all on all tables in schema private from %I', r);
      execute format('revoke all on all functions in schema private from %I', r);
      -- Public tables: read-only through RLS policies above.
      execute format('revoke insert, update, delete, truncate on public.players, public.rulesets, public.games, public.game_participants from %I', r);
    end if;
  end loop;
  if exists (select 1 from pg_roles where rolname = 'service_role') then
    grant usage on schema private to service_role;
    grant all on all tables in schema private to service_role;
    grant execute on all functions in schema private to service_role;
  end if;
end;
$$;
