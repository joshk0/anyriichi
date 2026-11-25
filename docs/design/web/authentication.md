# Authentication & platform account linking

Status: **Accepted** (D-003, D-009, D-010, D-011) · Related: [user model](../storage/user-model.md), [log privacy](log-submission-and-privacy.md)

## Provider — Supabase Auth
- Free to 50k MAU. Integrated with Postgres RLS (`auth.uid()`).
- Next.js via `@supabase/ssr` (cookie sessions in Server Components, Route Handlers, middleware).
- The Go worker never authenticates end users; it connects with a service role.

### V1 sign-in methods (D-010)
| Method | Notes |
|---|---|
| Google OAuth | |
| Email + password | with email confirmation & password reset |
| Discord OAuth | riichi community lives on Discord |

**Not in v1:** 2FA, X/LINE/Apple, linking/switching between sign-in methods (Q-020, Q-021).

## Linking platform accounts
- A user may link **at most one account per platform** (D-009).
- Linking creates a `player_claim` with status `unverified → pending → verified` (or `rejected`).
- Unverified claims still show stats on the user's page; privileged actions (privacy settings applied to that player, profile hiding — Q-010) require `verified`.

### Tenhou verification (D-011)
1. **25 ranked logs**: the claim auto-verifies when the claiming user has submitted 25 ranked-game logs containing that player name. No further conditions (D-022).
2. **ID secret** (optional, instant) — used once to prove ownership, then deleted ("Your login information will be deleted after verifying your account."). Mechanism TBD (Q-025).
3. **Manual review** — "This account is claimed by someone else" → operator reviews evidence, may transfer.
- Auto-verified claims can be **revoked** by a stronger proof (ID secret / review).

### Mahjong Soul verification (Q-023, proposal)
- Via the browser extension: the user's in-game session proves the account id → verified.
- Fallback: put a short nonce in the in-game signature; our fetcher reads it via `fetchAccountInfo`.

## Extension authentication
See [browser-extension](../extension/browser-extension.md) (Q-050).
