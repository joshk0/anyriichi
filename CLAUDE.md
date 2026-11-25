# anyRiichi — agent instructions

anyRiichi is a web app that aggregates riichi mahjong game records across platforms
(Mahjong Soul, Tenhou, later Riichi City) and shows player-centric statistics, in English and Japanese.

## Where things live
- `docs/design/` — the design spec hierarchy. **Start at `docs/design/README.md`.**
  - `docs/design/00-original-spec.md` — the owner's original spec, verbatim. Do not edit.
  - `docs/design/open-questions.md` — every unresolved question, with an ID (`Q-###`). Resolve them here when they become relevant, and record the outcome in `decisions.md`.
  - `docs/design/decisions.md` — decision log (`D-###`), append-only.
- `docs/research/` — raw research notes about reference sites, platform APIs and formats. Facts tagged [V] verified / [I] inferred.
- `memory/` — durable agent memory for this project (see below).
- Backlog: GitHub issues + GitHub Project on `joshk0/anyriichi`. Use the `gh` CLI. Initial seed: `scripts/backlog/backlog.json` + `create-backlog.sh` (idempotent).

## Working rules (from the owner)
- **Never build a design without confirmation.** Propose first (doc section, ASCII/HTML mockup, or PR), get explicit approval, then implement. Design docs carry a `Status:` line — only `Accepted` designs get implemented.
- Ask as many questions as needed; record unanswered ones in `open-questions.md` rather than guessing.
- Record memory in this repository (`memory/` or this file), **not** in the Claude account's private memory — the project may move between Claude accounts.
- Platform credentials and test game IDs live in `.env` (git-ignored). Never commit them, never paste them into issues/PRs/docs.
- Respect each platform's usage rules (e.g. Tenhou raw listing polled no more often than every 20 minutes).
- **Pre-prototype git workflow:** work in the local tree, let the diff build up, then commit and push directly to `main` (no small PRs). Once we reach a workable prototype, switch to targeted, readable PRs. Never push to a branch whose PR is already merged.

## Memory
`memory/MEMORY.md` is the index (one line per memory file). Each memory file holds one fact with
frontmatter (`name`, `description`, `type: user|feedback|project|reference`). Read the index at the
start of a session; update or delete stale entries rather than duplicating. Don't store what the code,
git history, or design docs already record.

## Tech stack (see docs/design/infrastructure/platform.md for status)
- Web: TypeScript, Next.js (App Router), Tailwind + shadcn/ui, hosted on Vercel — *accepted* (D-002, D-004)
- Auth + DB: Supabase (Auth + Postgres) — *accepted* (D-003)
- Worker: Go, standalone module, always-on GCE e2-micro (free tier) — *accepted* (D-004)
- Milestones: Tenhou-first MVP, then Mahjong Soul; Riichi City after v1 (D-005, D-017)

## UI mockups
Live on a private Claude design canvas (link in docs/design/web/user-interface.md) and are **mirrored in
`docs/design/web/mockups/`**. After any canvas change, sync the files back into the repo (see that README).
