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

### Detection patterns (initial; owner will supply more examples)
- Tenhou: `(\d{10}gm-[0-9a-f]{4}-\d{4,5}-x?[0-9a-f]{8,16})`, optionally inside `tenhou.net/0/?log=…&tw=N`.
- Majsoul: `paipu=` + `\d{6}-[0-9a-f-]{36}` (+ optional `_a\d+` seat/account suffix) on `game.mahjongsoul.com`, `mahjongsoul.game.yo-star.com`, `game.maj-soul.com`.

### Acceptance rules (D-012)
- **Ranked games only.** Tenhou: lobby `0000` and ranked rule types. Majsoul: rank-match mode ids (Bronze→Throne, 3p/4p, East/South). Others get an inline "Not ranked — not accepted" status.
- Duplicate → "Already imported" (still links the submitter for visibility purposes).
- The worker fetches the log itself from the platform; submissions are just IDs.

### Missing-logs list (D-019)
Games known from Tenhou public results (`scb`) for the user's claimed player that have no full log → listed with date/room/placement/table. Each row has a **Submit log** button opening a modal for that game's log URL; the submitted log must match the game (start hour, room, players). On success the row disappears and the next missing game is shown (D-025).

## Visibility (D-007, D-008)
Each user has a **log visibility** setting applied to logs they submitted / logs of their verified players:

| Setting | Who can open replay link & per-game detail |
|---|---|
| **Private** (default) | the user only |
| Players at the table | signed-in users with a claimed player in that game |
| Public | anyone |

Stats (aggregates) are always visible. Per-hand data of *other* players derived from a log is aggregated into their stats but the log itself follows the rules above.

Open: per-log overrides (Q-012); conflicting settings among table-mates (Q-013, proposal: strictest wins).
