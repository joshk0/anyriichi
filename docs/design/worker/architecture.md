# Worker architecture

Status: **Proposed** · Related: [platform](../infrastructure/platform.md)

- **Language:** Go (owner preference), standalone module in `/worker`, deployable on its own (single static binary + systemd, or a container).
- **Runtime:** one always-on process on a GCE e2-micro (D-004). 1 GB RAM / shared vCPU → keep memory bounded, stream gz files.
- **DB access:** direct Postgres connection to Supabase (pooler, session mode) using a dedicated role; `pgx` + `sqlc`.
- **Job model:** a Postgres-backed queue (`river` library or a simple `jobs` table with `SELECT … FOR UPDATE SKIP LOCKED`). No extra infra (no Redis). Web app enqueues jobs (e.g. log submission) by inserting rows; worker `LISTEN/NOTIFY` for latency.
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
