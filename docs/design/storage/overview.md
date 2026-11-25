# Storage overview

Status: **Proposed**

- **Postgres (Supabase)** for everything relational: users, players, games, participants, per-hand stats, caches.
- **Object storage** for raw logs (mjlog XML, Majsoul protobuf, tenhou.net/6 JSON) — private bucket, never publicly served (ToS, D-007). Choice pending (Q-041): Supabase Storage (simple, same vendor, egress from GCE counts) vs GCS (same cloud as worker, free 5 GB in us regions).
- **Migrations** via Supabase CLI in `/db/migrations`; types generated for TS (`supabase gen types`) and Go (sqlc).

## Volume estimate (Tenhou)
`scb` ≈ 1,300 games/hour observed → ~30k games/day → ~11M games/year, ~45M participant rows/year. At ~100 B/row ≈ 5 GB/yr incl. indexes → exceeds Supabase Free (500 MB) quickly → Pro ($25/mo, 8 GB) needed early in M1, or partition/archive older years (Q-071).

## Retention
- Keep all game summaries indefinitely.
- Keep raw logs for submitted/registered-user games; 鳳凰 public logs: store derived per-hand stats, raw optional.
