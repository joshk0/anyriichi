-- Postgres-backed job queue for the Go worker (docs/design/worker/architecture.md, D-043).
-- Lives in a private schema that is NOT exposed through the Supabase Data API.
-- The web app enqueues via queue.enqueue() from trusted server code only.

create schema if not exists queue;

create table queue.jobs (
  id            bigint generated always as identity primary key,
  kind          text        not null check (kind ~ '^[a-z0-9][a-z0-9_.-]*$'),
  payload       jsonb       not null default '{}'::jsonb,
  state         text        not null default 'pending'
                            check (state in ('pending', 'running', 'succeeded', 'failed', 'cancelled')),
  priority      smallint    not null default 0,
  run_at        timestamptz not null default now(),
  attempts      integer     not null default 0,
  max_attempts  integer     not null default 5 check (max_attempts > 0),
  -- Optional de-duplication key: at most one pending/running job per (kind, unique_key).
  unique_key    text,
  locked_by     text,
  locked_at     timestamptz,
  last_error    text,
  created_at    timestamptz not null default now(),
  updated_at    timestamptz not null default now(),
  finished_at   timestamptz
);

comment on table queue.jobs is 'Worker job queue; claimed with FOR UPDATE SKIP LOCKED.';

-- Claim path: next ready job of a kind.
create index jobs_ready_idx on queue.jobs (kind, priority desc, run_at, id) where state = 'pending';
-- Stale-lock rescue and pruning.
create index jobs_running_idx on queue.jobs (locked_at) where state = 'running';
create index jobs_finished_idx on queue.jobs (finished_at) where finished_at is not null;
create unique index jobs_unique_active_idx on queue.jobs (kind, unique_key)
  where unique_key is not null and state in ('pending', 'running');

-- Enqueue a job. Returns the new job id, or null when an active job with the same
-- (kind, unique_key) already exists. Notifies listening workers.
create function queue.enqueue(
  p_kind         text,
  p_payload      jsonb       default '{}'::jsonb,
  p_run_at       timestamptz default now(),
  p_priority     smallint    default 0,
  p_unique_key   text        default null,
  p_max_attempts integer     default 5
) returns bigint
language plpgsql
set search_path = ''
as $$
declare
  v_id bigint;
begin
  insert into queue.jobs (kind, payload, run_at, priority, unique_key, max_attempts)
  values (p_kind, coalesce(p_payload, '{}'::jsonb), coalesce(p_run_at, now()),
          coalesce(p_priority, 0), p_unique_key, coalesce(p_max_attempts, 5))
  on conflict (kind, unique_key) where unique_key is not null and state in ('pending', 'running')
  do nothing
  returning id into v_id;

  if v_id is not null then
    perform pg_notify('queue_jobs', p_kind);
  end if;
  return v_id;
end;
$$;

-- Lock everything down: only the table owner (postgres / service role paths) touches the queue.
alter table queue.jobs enable row level security;
revoke all on schema queue from public;
revoke all on all tables in schema queue from public;
revoke all on function queue.enqueue(text, jsonb, timestamptz, smallint, text, integer) from public;

do $$
declare r text;
begin
  foreach r in array array['anon', 'authenticated'] loop
    if exists (select 1 from pg_roles where rolname = r) then
      execute format('revoke all on schema queue from %I', r);
      execute format('revoke all on all tables in schema queue from %I', r);
      execute format('revoke all on function queue.enqueue(text, jsonb, timestamptz, smallint, text, integer) from %I', r);
    end if;
  end loop;
  if exists (select 1 from pg_roles where rolname = 'service_role') then
    grant usage on schema queue to service_role;
    grant select, insert, update, delete on queue.jobs to service_role;
    grant execute on function queue.enqueue(text, jsonb, timestamptz, smallint, text, integer) to service_role;
  end if;
end;
$$;
