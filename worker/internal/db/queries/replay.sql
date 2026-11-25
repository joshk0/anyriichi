-- name: QueueRankReplayForGames :exec
INSERT INTO private.rank_replay_queue (player_id)
SELECT DISTINCT gp.player_id FROM public.game_participants gp
WHERE gp.game_id = ANY(sqlc.arg(game_ids)::bigint[]) AND gp.player_id IS NOT NULL
ON CONFLICT (player_id) DO UPDATE SET queued_at = now();

-- name: QueueRankReplayAll :execrows
INSERT INTO private.rank_replay_queue (player_id)
SELECT id FROM public.players WHERE platform = 'tenhou'
ON CONFLICT (player_id) DO NOTHING;

-- name: ClaimRankReplay :many
-- Takes a batch of queued players; run inside the transaction that writes their results.
DELETE FROM private.rank_replay_queue
WHERE player_id IN (
  SELECT q.player_id FROM private.rank_replay_queue q ORDER BY q.queued_at LIMIT sqlc.arg(batch)::int
  FOR UPDATE SKIP LOCKED)
RETURNING player_id;

-- name: ListReplayGames :many
SELECT gp.player_id, gp.game_id, gp.placement, gp.rank_before, r.players, r.lobby, r.length
FROM public.game_participants gp
JOIN public.games g ON g.id = gp.game_id
JOIN public.rulesets r ON r.id = g.ruleset_id
WHERE gp.player_id = ANY(sqlc.arg(player_ids)::bigint[]) AND g.platform = 'tenhou'
ORDER BY gp.player_id, r.players, g.started_at, g.id;

-- name: WriteReplayResults :exec
-- Arrays are parallel; -99999 / '' mean unknown (null).
UPDATE public.game_participants gp SET
  est_rank_after = nullif(u.rank_after, -99999),
  est_pt_min     = nullif(u.pt_min, -99999),
  est_pt_max     = nullif(u.pt_max, -99999),
  rank_event     = nullif(u.event, ''),
  rank_pt_delta  = coalesce(nullif(u.delta, -99999), gp.rank_pt_delta)
FROM (SELECT unnest(sqlc.arg(game_ids)::bigint[]) AS game_id, unnest(sqlc.arg(placements)::smallint[]) AS placement,
             unnest(sqlc.arg(rank_after)::integer[]) AS rank_after, unnest(sqlc.arg(pt_min)::integer[]) AS pt_min,
             unnest(sqlc.arg(pt_max)::integer[]) AS pt_max, unnest(sqlc.arg(events)::text[]) AS event,
             unnest(sqlc.arg(deltas)::integer[]) AS delta) u
WHERE gp.game_id = u.game_id AND gp.placement = u.placement;
