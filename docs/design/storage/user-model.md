# User model

Status: **Proposed** (incorporates D-008, D-009, D-011) · Related: [authentication](../web/authentication.md)

Two distinct concepts:
- **User** — a person with an anyRiichi login (Supabase `auth.users`).
- **Player** — an identity on a game platform. Players exist whether or not anyone claims them.

```sql
profiles(
  user_id uuid PK references auth.users,
  handle text unique,                -- for /u/{handle}
  locale text check (locale in ('en','ja')),
  log_visibility text not null default 'private'  -- private | table | public (D-008, evaluated per participant page: D-029)
    check (log_visibility in ('private','table','public')),
  created_at timestamptz)

players(
  id bigserial PK,
  platform text check (platform in ('tenhou','majsoul','riichicity')),
  platform_key text,                 -- tenhou name | majsoul account_id | RC user id
  zone text null,                    -- majsoul CN/JP/EN (from account_id >> 23)
  display_name text,
  first_seen_at, last_seen_at timestamptz,
  unique(platform, platform_key))

player_names(player_id, name, first_seen_at, last_seen_at)   -- majsoul nickname history

player_claims(
  id bigserial PK,
  user_id uuid references profiles,
  player_id bigint references players,
  platform text,                     -- denormalized for the uniqueness rule below
  status text check (status in ('unverified','pending','verified','rejected','revoked')),
  method text null,                  -- 'log_count' | 'id_secret' | 'manual' | 'extension' | 'signature'
  opened_at timestamptz, verified_at timestamptz,
  unique(user_id, platform))         -- one account per platform (D-009)
-- partial unique index: one verified claim per player
create unique index on player_claims(player_id) where status = 'verified';

log_submissions(
  id bigserial PK, user_id uuid, platform text, external_game_id text,
  source text check (source in ('paste','extension','premium_import')),
  status text,  -- queued | fetched | rejected_not_ranked | duplicate | error
  error text, game_id bigint null, created_at timestamptz,
  unique(user_id, platform, external_game_id))
```

- Tenhou verification counter = count of `log_submissions` by the user whose game includes the claimed player, ranked, played after `opened_at` (D-011/Q-022).
- RLS: `profiles`, `player_claims`, `log_submissions` owner-only; `players` public read; worker writes via service role.
