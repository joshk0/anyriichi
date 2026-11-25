-- name: GetWorkerState :one
SELECT value FROM private.worker_state WHERE key = $1;

-- name: SetWorkerState :exec
INSERT INTO private.worker_state (key, value, updated_at) VALUES ($1, $2, now())
ON CONFLICT (key) DO UPDATE SET value = excluded.value, updated_at = now();

-- name: GetRawFileSizes :many
SELECT file, size FROM private.tenhou_raw_files WHERE file = ANY(sqlc.arg(files)::text[]);

-- name: UpsertRawFile :exec
INSERT INTO private.tenhou_raw_files (file, size, lines, processed_at) VALUES ($1, $2, $3, now())
ON CONFLICT (file) DO UPDATE SET size = excluded.size, lines = excluded.lines, processed_at = now();

-- name: UpsertTenhouSummary :batchone
SELECT private.upsert_tenhou_summary(
  sqlc.arg(match_key)::text, sqlc.arg(started_at)::timestamptz, sqlc.arg(duration)::smallint,
  sqlc.arg(source)::text, sqlc.arg(code)::text, sqlc.arg(players)::smallint, sqlc.arg(length)::text,
  sqlc.arg(lobby)::text, sqlc.arg(aka)::boolean, sqlc.arg(kuitan)::boolean, sqlc.arg(speed)::text,
  sqlc.arg(names)::text[], sqlc.arg(pt_tenths)::integer[], sqlc.arg(rank_deltas)::integer[]
)::bigint AS game_id;

-- name: QueueTenhouLog :one
-- Registers a log id (idempotent) and enqueues a fetch job when it is new.
-- Returns true when the log was newly registered.
WITH ins AS (
  INSERT INTO private.tenhou_logs (log_id, source) VALUES (sqlc.arg(log_id)::text, sqlc.arg(source)::text)
  ON CONFLICT (log_id) DO NOTHING
  RETURNING log_id
), q AS (
  SELECT queue.enqueue('tenhou.log', jsonb_build_object('log_id', log_id), now(),
                       sqlc.arg(priority)::smallint, log_id, 5)
  FROM ins
)
SELECT EXISTS (SELECT 1 FROM q)::boolean AS queued;

-- name: GetTenhouLog :one
SELECT log_id, status, game_id, source FROM private.tenhou_logs WHERE log_id = $1;

-- name: MarkTenhouLog :exec
UPDATE private.tenhou_logs
SET status = sqlc.arg(status)::text, reject_reason = sqlc.narg(reject_reason)::text,
    raw_gz = coalesce(sqlc.narg(raw_gz)::bytea, raw_gz), game_id = coalesce(sqlc.narg(game_id)::bigint, game_id),
    fetched_at = CASE WHEN sqlc.narg(raw_gz)::bytea IS NULL THEN fetched_at ELSE now() END,
    updated_at = now()
WHERE log_id = sqlc.arg(log_id)::text;

-- name: UpsertTenhouFullLog :one
SELECT private.upsert_tenhou_full_log(
  sqlc.arg(match_key)::text, sqlc.arg(started_at)::timestamptz, sqlc.arg(code)::text,
  sqlc.arg(players)::smallint, sqlc.arg(length)::text, sqlc.arg(lobby)::text, sqlc.arg(aka)::boolean,
  sqlc.arg(kuitan)::boolean, sqlc.arg(speed)::text, sqlc.arg(names)::text[], sqlc.arg(pt_tenths)::integer[],
  sqlc.arg(rank_deltas)::integer[], sqlc.arg(seats)::integer[], sqlc.arg(scores)::integer[],
  sqlc.arg(dans)::integer[], sqlc.arg(rates)::float8[]::numeric[]
)::bigint AS game_id;

-- name: DeleteHands :exec
DELETE FROM private.hands WHERE game_id = $1;

-- name: InsertHand :exec
INSERT INTO private.hands (game_id, hand_index, round, honba, riichi_sticks, dealer_seat, outcome)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: InsertHandPlayerStats :exec
INSERT INTO private.hand_player_stats (
  game_id, hand_index, seat, player_id, won, won_by, dealt_in, riichi, called, tenpai_at_draw,
  han, fu, win_value, deal_in_value, yaku, score_delta)
VALUES (
  sqlc.arg(game_id), sqlc.arg(hand_index), sqlc.arg(seat),
  (SELECT gp.player_id FROM public.game_participants gp WHERE gp.game_id = sqlc.arg(game_id) AND gp.seat = sqlc.arg(seat)),
  sqlc.arg(won), sqlc.narg(won_by), sqlc.arg(dealt_in), sqlc.arg(riichi), sqlc.arg(called), sqlc.narg(tenpai_at_draw),
  sqlc.narg(han), sqlc.narg(fu), sqlc.narg(win_value), sqlc.narg(deal_in_value), sqlc.arg(yaku), sqlc.arg(score_delta));

-- name: ListStoredTenhouLogs :many
SELECT log_id, raw_gz FROM private.tenhou_logs WHERE status = 'stored' AND raw_gz IS NOT NULL ORDER BY log_id;
