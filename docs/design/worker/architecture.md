# Worker architecture

Status: **Accepted** (queue + scheduler implemented in M0, #7; ingest jobs Proposed) · Related: [platform](../infrastructure/platform.md)

- **Language:** Go (owner preference), standalone module in `/worker`, deployable on its own (single static binary + systemd, or a container).
- **Runtime:** one always-on process on a GCE e2-micro (D-004). 1 GB RAM / shared vCPU → keep memory bounded, stream gz files.
- **DB access:** direct Postgres connection to Supabase (pooler, session mode, pool ≤ 6) with `pgx` + `sqlc`. Currently the `postgres` role; a dedicated least-privilege role is a follow-up before launch.
- **Job model (implemented, D-043):** own `queue.jobs` table claimed with `SELECT … FOR UPDATE SKIP LOCKED` (not river — the web app must enqueue from TypeScript, and river's schema is Go-client-specific). No extra infra (no Redis).
  - Schema `queue` is private: not exposed via the Supabase Data API; `anon`/`authenticated` have no access; `service_role` may enqueue. Migration: `supabase/migrations/20261125200000_job_queue.sql`.
  - Enqueue with `select queue.enqueue(kind, payload, run_at, priority, unique_key, max_attempts)` — returns the id, or null when an active job with the same `(kind, unique_key)` exists; sends `NOTIFY queue_jobs, kind`.
  - Worker (`worker/internal/queue`): per-kind claim loops (`Concurrency` = politeness limit, e.g. 1 for tenhou.net), per-job timeout, retries with exponential backoff (10 s → 1 h, default 5 attempts), `queue.Permanent(err)` for non-retryable failures, panic recovery. One hijacked connection `LISTEN`s (works through the Supabase session pooler); 30 s poll fallback.
  - Maintenance (scheduler, every 5 min): rescue `running` jobs locked > 15 min (crashed worker); prune succeeded jobs after 7 days. Failed jobs are kept for inspection.
  - Queries are typed with sqlc (`worker/internal/db`, `make sqlc`); CI checks the generated code is current and runs DB tests against a Postgres 17 service.
  - Smoke test: `select queue.enqueue('system.ping', '{"from":"me"}');` → logged by the worker.
- **Scheduler:** in-process cron (`robfig/cron` or river periodic jobs):
  | Job | Cadence |
  |---|---|
  | Tenhou `list.cgi` poll + changed-file fetch | every 20 min (hard minimum, ToS) |
  | Tenhou mjlog fetch (submitted / scc) | continuous, rate-limited (~1 req / 2 s), single concurrency |
  | AK sync for registered Majsoul players | hourly (M2) |
  | Majsoul `fetchGameRecord` for registered users | continuous, rate-limited (M2) |
  | Stats-cache invalidation / recompute | on ingest |
  | Percentile distributions | nightly |
- **Per-host politeness:** a shared rate limiter per external host; one concurrent session for tenhou.net.
- **Observability:** structured logs (slog) → Cloud Logging (free tier); `ingest_runs` table with counts/errors for an operator page; health endpoint; alert on stalled polling.
- **Config/secrets:** env vars from GCP Secret Manager (free tier: 6 secrets) or a root-only env file on the VM; platform credentials never in the repo.
- **Deploy:** GitHub Actions builds the binary, `gcloud compute scp` + `systemctl restart` (or container-optimized OS). See platform doc.
