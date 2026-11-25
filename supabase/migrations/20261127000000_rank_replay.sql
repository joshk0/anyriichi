-- Tenhou dan/pt replay (#19, docs/design/worker/tenhou-ingest.md "Dan / pt replay").
-- Results are public like the rest of game_participants (rank is public on Tenhou).

alter table public.game_participants
  add column est_rank_after smallint,                       -- rank after this game, when known
  add column est_pt_min     integer,                        -- pt range after this game (rank est_rank_after)
  add column est_pt_max     integer,
  add column rank_event     text check (rank_event in ('promotion', 'demotion'));

-- Players whose games changed since their last replay. Filled by the worker after each ingest.
create table private.rank_replay_queue (
  player_id bigint primary key references public.players on delete cascade,
  queued_at timestamptz not null default now()
);

-- Replay everyone already ingested.
insert into private.rank_replay_queue (player_id) select id from public.players where platform = 'tenhou';

revoke all on private.rank_replay_queue from public;
do $$
declare r text;
begin
  foreach r in array array['anon', 'authenticated'] loop
    if exists (select 1 from pg_roles where rolname = r) then
      execute format('revoke all on private.rank_replay_queue from %I', r);
    end if;
  end loop;
  if exists (select 1 from pg_roles where rolname = 'service_role') then
    grant all on private.rank_replay_queue to service_role;
  end if;
end;
$$;
