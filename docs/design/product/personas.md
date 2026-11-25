# Personas

Status: **Proposed** · Owner: — · Related: [user-interface](../web/user-interface.md), [roadmap](roadmap.md)

anyRiichi targets English- and Japanese-speaking riichi mahjong players who play on online platforms.

## P1 — The Ladder Climber (primary)
- Plays ranked daily on 1–2 platforms (e.g. Mahjong Soul Jade + Tenhou 特上).
- Wants: placement rates, average rank, rank-point trajectory, "am I improving?", stable-rank estimate.
- Today uses Amae Koromo and/or Nodocchi separately; no single view across platforms.
- Device: desktop and phone roughly equal. Mostly checks stats on phone after a session.

## P2 — The Improver / Analyst
- Studies the game; uses NAGA, Mortal (mjai.ekyu.moe), Bigcoach.
- Wants: detailed per-hand stats (deal-in rate, riichi stats, call rate, win shape), comparisons against rank peers, one-click "send this game to AI review" regardless of platform.
- Will install a browser extension and bulk-upload logs if it unlocks richer stats (esp. Tenhou).

## P3 — The Casual / Multi-platform player
- Plays a few games a week, on whatever platform friends use; may play mostly on mobile.
- Wants: "how am I doing overall?", simple placement chart, recent games. Low tolerance for setup.
- Cannot use the browser extension (mobile) → needs bulk paste of log URLs.

## P4 — The Spectator / Community member (anonymous)
- Not registered. Looks up a streamer, a friend, or a top player; checks leaderboards.
- Wants: fast, shareable player pages with good link previews (Discord/X), JP/EN.
- Only served for data we ingest publicly (Tenhou all public games; Majsoul via upstream — see D-001).

## P5 — Club / League organiser (later)
- Runs a friend-room league or tournament on Majsoul/Tenhou/RC.
- Wants: group a set of games/players into a "league", standings table, export.
- Out of scope for MVP; captured as a later epic.

## P6 — Operator (us)
- Needs: ingestion health dashboards, re-processing tools, account linking disputes, abuse handling, takedown/opt-out requests.

## Open questions
- ~~Q-010~~ Resolved → D-030: no profile hiding.
- Q-011: Is P5 (leagues) in scope for v1?
