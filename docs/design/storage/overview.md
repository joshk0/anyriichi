# Storage overview

Status: **Proposed**

- **Postgres (Supabase)** for everything relational: users, players, games, participants, per-hand stats, caches.
- **Object storage** for raw logs (mjlog XML, Majsoul protobuf, tenhou.net/6 JSON) — private bucket, never publicly served (ToS, D-007). Supabase Storage private bucket (D-039).
- **Migrations** via Supabase CLI in `/db/migrations`; types generated for TS (`supabase gen types`) and Go (sqlc).

## Volume estimate (Tenhou)
`scb` ≈ 1,300 games/hour observed → ~30k games/day → ~11M games/year, ~45M participant rows/year. At ~100 B/row ≈ 5 GB/yr incl. indexes → Supabase Pro from M1 (D-039).

## Retention
- Keep all game summaries indefinitely.
- Keep raw logs for submitted/registered-user games; 鳳凰 public logs: store derived per-hand stats, raw optional.
