# Web application architecture

Status: **Accepted** (scaffolded in M0, D-041) · Related: [platform](../infrastructure/platform.md), [storage overview](../storage/overview.md)

## Stack
- Next.js 16 (App Router, `cacheComponents`, Proxy = former middleware) + TypeScript, deployed on Vercel (D-002, D-004). **Read `apps/web/node_modules/next/dist/docs/` before writing Next code** — APIs differ from older versions (see `apps/web/AGENTS.md`).
- UI: Tailwind CSS v4 with design tokens in `globals.css` `@theme`; shadcn/ui (Radix primitives) to be added when the first interactive components land; charts: Recharts or visx (Q: pick during M1 — candlestick support needed for rank charts).
- Data: Supabase Postgres via `@supabase/ssr` (RLS for user data) + server-only service queries for heavy stats.
- i18n: Next.js built-in pattern — `app/[lang]` root segment, `proxy.ts` redirect by `Accept-Language`, JSON dictionaries loaded via `next/root-params` (no `next-intl`, D-041); locales `en`, `ja`; locale in URL prefix (`/ja/player/...`).
- Validation: zod. Testing: Vitest + Playwright.

## Routes (proposed)
| Route | Purpose |
|---|---|
| `/` | search + recent activity |
| `/p/{platform}/{playerKey}` | single platform player page (public) |
| `/u/{handle}` | registered user's combined cross-platform page |
| `/submit` | log submission ([doc](log-submission-and-privacy.md)) |
| `/settings/*` | linked accounts, log privacy, sign-in, extension |
| `/api/*` | route handlers: submission, extension API, search |

Filter state is encoded in the URL query (`?platform=tenhou&players=4&lobby=tokujou&len=south&type=ranked&from=…&to=…`) so pages are shareable (AK pattern).

## Rendering & performance
- Player pages: server-rendered with cached aggregates (`stats_cache` table keyed by player + filter hash; invalidated by the worker on new games). OG image generation for link previews (persona P4).
- Heavy per-filter aggregates computed by SQL over `game_participants` / `hand_stats`; precomputed rollups for common filters (see [statistics model](../storage/statistics-model.md)).

## Monorepo layout (proposed)
```
/apps/web          Next.js app
/apps/extension    browser extension (TS, WXT or Plasmo)
/worker            Go module (standalone deployable)
/packages/shared   TS: log-id detection, rule tables, stat formulas, types generated from DB schema
/supabase          Supabase CLI config + SQL migrations, seeds
/docs              design + research
```
Stat formulas & rule tables are needed in both Go (worker) and TS (web) → keep tables as JSON data in `/packages/shared/rules/` consumed by both.
