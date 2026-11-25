# Infrastructure & platform

Status: **Accepted** (D-002, D-003, D-004) / **Proposed** (details)

| Concern | Choice | Cost (est.) |
|---|---|---|
| Web hosting | **Vercel** (Hobby → Pro if commercial) | $0 → $20/mo |
| Auth | **Supabase Auth** | $0 (≤50k MAU) |
| Database | **Supabase Pro** from M1, us-east-1 (D-039; analysis: docs/research/db-cost-analysis.md) | $25/mo → ~$31 by Y3 |
| Worker compute | **GCE e2-micro** Always Free (us-west1/us-central1/us-east1), 30 GB pd-standard | $0 (1 GB/mo free egress; watch DB-write egress) |
| Scheduling | in-process scheduler in the Go worker (no Cloud Scheduler needed) | $0 |
| Queue | Postgres-backed (river / SKIP LOCKED) | $0 |
| Raw log storage | Supabase Storage, private bucket (100 GB incl. in Pro) | $0 |
| Secrets | GCP Secret Manager / VM env file; Vercel env vars; local `.env` (git-ignored) | $0 |
| DNS/domain | Q-070 | ~$12/yr |
| CI/CD | GitHub Actions: lint/test/build; Vercel Git integration for web previews per PR; worker deploy job on `main` | $0 |
| Monitoring | Vercel analytics (free), Cloud Logging, Sentry free tier (optional) | $0 |

Region alignment: put Supabase in AWS `us-west-1`/`us-east-1` near the GCE VM region to minimise latency; JP users are served by Vercel's edge for static/ISR content.

## Key libraries
- Web: next, react, @supabase/ssr, @supabase/supabase-js, tailwindcss, shadcn/ui, next-intl, zod, recharts/visx, vitest, playwright.
- Worker: Go ≥ 1.23, pgx, sqlc, river (or custom queue), robfig/cron, gorilla/websocket or nhooyr/websocket, google.golang.org/protobuf, slog.
- Shared: JSON rule tables, log-id regexes (TS + Go implementations, shared test vectors).

## Agentic-development affordances
- Vercel preview URL per PR; Supabase branching (Pro) or a separate dev project for migrations testing.
- `make`/`pnpm` scripts for everything; seed data + fixtures so agents can run locally without platform credentials.
