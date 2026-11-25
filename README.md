# anyRiichi

Riichi mahjong stats across platforms (Tenhou first, then Mahjong Soul), in English and Japanese.

- Design specs: [`docs/design/README.md`](docs/design/README.md) · decisions: [`docs/design/decisions.md`](docs/design/decisions.md)
- Backlog: [GitHub Project](https://github.com/users/joshk0/projects/1)
- Agent instructions: [`CLAUDE.md`](CLAUDE.md)

## Layout
| Path | What |
|---|---|
| `apps/web` | Next.js 16 app (App Router, Tailwind v4, `[lang]` i18n routing: `en`, `ja`) |
| `packages/shared` | TS shared code: log-reference detection, Tenhou rule decoding; shared test vectors in `testdata/` |
| `worker` | Go ingestion worker (standalone module) |
| `supabase` | Supabase CLI config + SQL migrations |
| `docs` | design specs, research, mockup mirror |
| `scripts/backlog` | backlog seed + creation script |

## Develop
Requires Node ≥ 22 (pnpm comes via corepack — no global install) and Go ≥ 1.25.

```sh
make install      # pnpm install
make dev          # web app on http://localhost:3000
make test         # TS + Go tests (shared vectors run in both)
make check        # typecheck + lint + test
make worker-run   # run the worker locally
```
Copy `.env.example` to `.env` for local secrets.
