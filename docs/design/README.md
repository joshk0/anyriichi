# anyRiichi design specs

Start here. Each doc has a `Status:` line — **Proposed** (needs owner sign-off), **Accepted** (build it), **Superseded**.
Implementation only proceeds on Accepted designs (see `/CLAUDE.md`).

- [00-original-spec.md](00-original-spec.md) — the owner's original brief (verbatim, frozen)
- [decisions.md](decisions.md) — decision log (D-###)
- [open-questions.md](open-questions.md) — unresolved questions (Q-###)

## Product
- [product/personas.md](product/personas.md) — who we build for
- [product/related-projects.md](product/related-projects.md) — Amae Koromo, Nodocchi, converters, AI reviewers (summary; raw notes in `/docs/research/`)
- [product/roadmap.md](product/roadmap.md) — milestones M0–M4

## Web application
- [web/architecture.md](web/architecture.md) — Next.js app structure, data access, rendering
- [web/authentication.md](web/authentication.md) — Supabase Auth, platform account linking & verification
- [web/user-interface.md](web/user-interface.md) — screens, mockups ([canvas](https://claude.ai/artifact/7Xck63BrfEFUchViwVWgRv) · [repo mirror](web/mockups/))
- [web/log-submission-and-privacy.md](web/log-submission-and-privacy.md) — submitting logs, visibility rules, ToS constraints

## Storage
- [storage/overview.md](storage/overview.md) — Postgres layout, raw-log storage, retention
- [storage/user-model.md](storage/user-model.md) — users, players, claims
- [storage/game-record-model.md](storage/game-record-model.md) — games, participants, rulesets, raw logs
- [storage/statistics-model.md](storage/statistics-model.md) — stat definitions, normalization, aggregation

## Worker
- [worker/architecture.md](worker/architecture.md) — Go worker, job scheduling, queues
- [worker/tenhou-ingest.md](worker/tenhou-ingest.md) — sc/raw polling, log fetching
- [worker/majsoul-ingest.md](worker/majsoul-ingest.md) — Amae Koromo + own record fetcher
- [worker/log-conversion.md](worker/log-conversion.md) — parsers, tenhou.net/6 conversion, AI-review export

## Browser extension
- [extension/browser-extension.md](extension/browser-extension.md)

## Infrastructure
- [infrastructure/platform.md](infrastructure/platform.md) — hosting, DB, scheduling, CI/CD, SDKs, costs
