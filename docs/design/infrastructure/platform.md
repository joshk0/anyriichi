# Infrastructure & platform

Status: **Accepted** (D-002, D-003, D-004) / **Proposed** (details)

| Concern | Choice | Cost (est.) |
|---|---|---|
| Web hosting | **Vercel** (Hobby → Pro if commercial) | $0 → $20/mo |
| Auth | **Supabase Auth** | $0 (≤50k MAU) |
| Database | **Supabase Pro** from M1, us-east-1 (D-039; analysis: docs/research/db-cost-analysis.md) | $25/mo → ~$31 by Y3 |
| Worker compute | **GCE e2-micro** Always Free, us-east1-b, 30 GB pd-standard, Standard network tier | $0 VM + ~$3.65/mo ephemeral external IPv4 (D-042) |
| Scheduling | in-process scheduler in the Go worker (no Cloud Scheduler needed) | $0 |
| Queue | Postgres-backed (river / SKIP LOCKED) | $0 |
| Raw log storage | Supabase Storage, private bucket (100 GB incl. in Pro) | $0 |
| Secrets | GCP Secret Manager / VM env file; Vercel env vars; local `.env` (git-ignored) | $0 |
| DNS/domain | Q-070 | ~$12/yr |
| CI/CD | GitHub Actions: lint/test/build; Vercel Git integration for web previews per PR; worker deploy job on `main` | $0 |
| Monitoring | Vercel analytics (free), Cloud Logging, Sentry free tier (optional) | $0 |

Region alignment: put Supabase in AWS `us-west-1`/`us-east-1` near the GCE VM region to minimise latency; JP users are served by Vercel's edge for static/ISR content.

## Provisioned resources (M0)
All on the owner's **personal** accounts (D-035). Never use the work GCP account / `ternary-dev`.

| Resource | Where | Notes |
|---|---|---|
| Vercel project `anyriichi` | team `joshk0` (Hobby), Git-connected to `joshk0/anyriichi`, root `apps/web`, build region iad1 | https://anyriichi.vercel.app · env: `NEXT_PUBLIC_SUPABASE_URL`, `NEXT_PUBLIC_SUPABASE_ANON_KEY` (all environments) |
| Supabase project `anyriichi` | org "joshk0's Org", us-east-1, ref in `supabase/.temp` after `supabase link` | Free plan during M0; upgrade to Pro at M1 (D-039). DB password + pooler URL in local `.env` only |
| GCP project `anyriichi` | billing enabled | APIs: compute, iap |
| VM `anyriichi-worker` | us-east1-b, e2-micro, Debian 13, Shielded VM, OS Login, no service account | ephemeral external IP (egress only). Firewall: SSH from IAP range `35.235.240.0/20` only (default ssh/rdp rules removed). Worker: systemd unit `anyriichi-worker`, binary `/opt/anyriichi/worker`, env `/etc/anyriichi/worker.env` (root:anyriichi 640), health on `localhost:8080/healthz` (not exposed) |

Deploy the worker: `make worker-deploy` (add `ENV=1` to also push worker keys from `.env`). Uses `gcloud compute scp/ssh --tunnel-through-iap`; refuses to run if gcloud is on a work account. A GitHub Actions deploy (Workload Identity Federation) can replace this later.

## Key libraries
- Web: next, react, @supabase/ssr, @supabase/supabase-js, tailwindcss, shadcn/ui, next-intl, zod, recharts/visx, vitest, playwright.
- Worker: Go ≥ 1.23, pgx, sqlc, river (or custom queue), robfig/cron, gorilla/websocket or nhooyr/websocket, google.golang.org/protobuf, slog.
- Shared: JSON rule tables, log-id regexes (TS + Go implementations, shared test vectors).

## Agentic-development affordances
- Vercel preview URL per PR; Supabase branching (Pro) or a separate dev project for migrations testing.
- `make`/`pnpm` scripts for everything; seed data + fixtures so agents can run locally without platform credentials.
