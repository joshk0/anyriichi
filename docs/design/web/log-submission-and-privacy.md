# Log submission & privacy

Status: **Accepted** (D-007, D-008, D-012, D-013, D-015, D-019) with open sub-questions Q-012, Q-013

## Constraints
- **Tenhou ToS**: no redistribution of log files/index; no service where the general public downloads Tenhou logs. Derived stats are fine (Nodocchi model). → We **never expose log links publicly by default**; replays open on tenhou.net, never served by us.
- Amae Koromo also masks game links (via `view_game` redirect). We follow the same principle for Majsoul.

## Submission channels
| Channel | Platforms | Notes |
|---|---|---|
| Paste box (one per line, or one at a time on mobile) | Tenhou, Majsoul | Platform auto-detected (D-013) |
| Browser extension | Tenhou, Majsoul | [extension doc](../extension/browser-extension.md) |
| Tenhou premium import | Tenhou | `tenhou.net/mjlog.html` by Tenhou ID (D-015, Q-033) |
| Worker (public feeds) | Tenhou 鳳凰 (`scc`) | ingested automatically, still subject to visibility rules |

### Detection patterns (validated against `docs/research/sample-logs.md`)
- **Tenhou — match the log ID only, ignore the URL.** Viewer paths vary (`/0/`, `/3/`, `/6/`, …) and hosts/params change, so
  detection never looks at the URL. If this regex matches anywhere in the input, it **is** a Tenhou log (bare ID or inside any URL):
  ```
  (?<![0-9A-Za-z])(\d{10}gm-[0-9a-f]{4}-\d{4,5}-x?[0-9a-f]{8,16})(?![0-9A-Za-z])
  ```
  Parts: `YYYYMMDDHH` (JST) · `gm` · 4-hex rule code · 4–5 digit lobby (`0000` ranked, `0001` L-lobby, `1xxxx` C-lobby) ·
  8-hex hash (or `x` + 12-hex obfuscated hash on older client IDs). An optional `tw=N` elsewhere in the input is read separately as the seat hint.
  The rule code tells players/length/lobby/speed before fetching, e.g. `0099` = 3p 南 上級, `0019` = 3p 南 一般, `0001` = 4p 東 一般,
  `00c1` = 4p 東 上級 **速** (0x40 speed bit — speed lobbies are ranked too).
  Test vectors: every Tenhou line in `docs/research/sample-logs.md` (+ bare IDs, IDs inside prose, `/6/` URLs, `x`-prefixed hashes; must not match when followed by extra alphanumerics). Shared between TS and Go via `/packages/shared` test vectors.
- **Mahjong Soul:** `paipu=YYMMDD-<uuid v4>` optionally followed by `_a<digits>` (the account/seat suffix is variable length, e.g. `_a9249`, `_a924963243`). Hosts: `mahjongsoul.game.yo-star.com`, `game.mahjongsoul.com`, `game.maj-soul.com`.
- **Reject Majsoul anonymised share links:** the "anonymous share" option scrambles the id into non-hex letters (e.g. `joklnq-v0zwv089-a480-6aa5-fgbo-lpnqsjsqro12_a924`). Detect = first segment not 6 digits / uuid not hex → "Anonymous share links aren't supported — use the normal share link."
- Room/ranked status for Majsoul is only known after fetching the record (friendly and tournament samples must be rejected post-fetch, D-012).

### Acceptance rules (D-012)
- **Ranked games only.** Tenhou: lobby `0000` and ranked rule types. Majsoul: rank-match mode ids (Bronze→Throne, 3p/4p, East/South). Others get an inline "Not ranked — not accepted" status.
- Duplicate → "Already imported" (still links the submitter for visibility purposes).
- The worker fetches the log itself from the platform; submissions are just IDs.

### Missing-logs list (D-019)
Games known from Tenhou public results (`scb`) for the user's claimed player that have no full log → listed with date/room/placement/table. Each row has a **Submit log** button opening a modal for that game's log URL; the submitted log must match the game (start hour, room, players). On success the row disappears and the next missing game is shown (D-025).

## Visibility (D-007, D-008, D-029)
Each user has one account-wide **log visibility** setting (no per-log overrides in v1):

| Setting | Who can reach the log (replay link & per-game detail) **from this user's pages** |
|---|---|
| **Private** (default) | the user only |
| Players at the table | signed-in users with a claimed player in that game |
| Public | anyone |

**Association:** a log submitted by anyone is automatically associated with every participant who has, or later binds, a
claimed account — exactly as if each had submitted it themselves.

**Per-page evaluation:** whether a log is linked is decided separately for each participant's page using *that*
participant's setting. Example: A (Public) and B (Private) played together; A's page links the log for everyone, B's
page does not. A's choice never exposes the log through B's page; only B's experience differs. Direct log URLs are
unguessable IDs; access check = "viewer is allowed by at least one participant whose page they came from".

Stats (aggregates) are always visible, and there is **no profile hiding** (D-030).
