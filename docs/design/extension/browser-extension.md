# Browser extension

Status: **Proposed** (scope accepted: D-014) · Milestone M3

## Goals
- **Tenhou** (tenhou.net web client): read localStorage ring buffer `lognext`, `log0`…`log39` (JSON with `log`, `oya`, `uname`, `sc`) → submit new ranked log ids; watch for new entries after each game.
- **Mahjong Soul** (game.mahjongsoul.com / yo-star / maj-soul.com): use the logged-in session (access token) to list the user's own paipu (`fetchGameRecordList`/`V2`) → bulk-submit history once, then incremental sync. Also proves Majsoul account ownership (verification, Q-023).

## Design
- Manifest V3, TypeScript, cross-browser via WXT (or Plasmo); Chrome Web Store + Firefox AMO (Q-051).
- Content scripts per site; background service worker submits to `POST /api/ext/submissions`.
- Auth to anyRiichi: OAuth-style connect flow issuing a scoped personal token stored in extension storage (Q-050).
- Minimal permissions: host permissions for the two game sites + anyRiichi only. Never send the platform token to anyRiichi servers — only log ids and the verified account id.
- Privacy: explain exactly what is read; open-source the extension code.

## Open
Q-050 auth, Q-051 store accounts, Q-052 Majsoul API details & ToS of in-page use.
