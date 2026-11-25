-- name: Enqueue :one
-- Returns 0 when an active job with the same (kind, unique_key) already exists.
SELECT coalesce(queue.enqueue(
  sqlc.arg(kind)::text,
  sqlc.arg(payload)::jsonb,
  sqlc.arg(run_at)::timestamptz,
  sqlc.arg(priority)::smallint,
  sqlc.narg(unique_key)::text,
  sqlc.arg(max_attempts)::integer
), 0)::bigint AS id;

-- name: ClaimJob :one
-- Atomically claims the next ready job of the given kind for this worker.
UPDATE queue.jobs j
SET state = 'running', attempts = j.attempts + 1, locked_by = sqlc.arg(locked_by)::text,
    locked_at = now(), updated_at = now()
WHERE j.id = (
  SELECT c.id FROM queue.jobs c
  WHERE c.state = 'pending' AND c.kind = sqlc.arg(kind)::text AND c.run_at <= now()
  ORDER BY c.priority DESC, c.run_at, c.id
  FOR UPDATE SKIP LOCKED
  LIMIT 1
)
RETURNING j.*;

-- name: CompleteJob :exec
UPDATE queue.jobs
SET state = 'succeeded', finished_at = now(), updated_at = now(), locked_by = NULL, locked_at = NULL, last_error = NULL
WHERE id = $1 AND state = 'running';

-- name: FailJob :exec
-- Reschedules the job after backoff, or marks it failed once attempts are exhausted
-- (or when permanent is true).
UPDATE queue.jobs
SET state = CASE WHEN sqlc.arg(permanent)::boolean OR attempts >= max_attempts THEN 'failed' ELSE 'pending' END,
    run_at = now() + make_interval(secs => sqlc.arg(backoff_seconds)::float8),
    finished_at = CASE WHEN sqlc.arg(permanent)::boolean OR attempts >= max_attempts THEN now() END,
    last_error = sqlc.arg(last_error)::text,
    locked_by = NULL, locked_at = NULL, updated_at = now()
WHERE id = sqlc.arg(id) AND state = 'running';

-- name: RescueStaleJobs :execrows
-- Returns jobs whose worker died mid-run to the pending state.
UPDATE queue.jobs
SET state = 'pending', locked_by = NULL, locked_at = NULL, updated_at = now(),
    last_error = 'lock expired'
WHERE state = 'running' AND locked_at < now() - make_interval(secs => sqlc.arg(older_than_seconds)::float8);

-- name: PruneFinishedJobs :execrows
DELETE FROM queue.jobs
WHERE state IN ('succeeded', 'cancelled') AND finished_at < now() - make_interval(secs => sqlc.arg(older_than_seconds)::float8);

-- name: GetJob :one
SELECT * FROM queue.jobs WHERE id = $1;
