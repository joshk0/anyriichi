# Open questions

Resolve a question when it becomes relevant for implementation; move the answer into `decisions.md` and mark it
**Resolved → D-xxx** here (keep the row for history). Owner = who must answer.

| ID | Area | Question | Owner | Status |
|---|---|---|---|---|
| Q-010 | Privacy | Can a player **hide their profile** (opt-out) even though the data is public on the source site? Takedown process? | Owner | Open |
| Q-011 | Product | Are leagues/clubs (persona P5) in scope for v1? (assumed no) | Owner | Open |
| Q-012 | Privacy | Log visibility: account-wide only, or **per-log override** (e.g. make one game public)? | Owner | Open |
| Q-013 | Privacy | When table-mates choose different visibility, does the **strictest** setting among claimed participants win, or each uploader's choice? (Tenhou logs show all 4 hands.) | Owner | Open |
| Q-020 | Auth | LINE login for JP users (custom OIDC)? (post-v1) | Owner | Deferred |
| Q-021 | Auth | Apple login (needs $99/yr Apple dev account)? (post-v1) | Owner | Deferred |
| Q-022 | Auth | Tenhou verification tightening: require logs played after claim start + spread over ≥5 days? Exact numbers? | Owner | Proposed (D-011) |
| Q-023 | Auth | Majsoul verification: extension-based (token proves account) vs. signature-nonce via `fetchAccountInfo`? Both? | Owner | Open |
| Q-024 | Auth | Can two users hold **unverified** claims on the same player? (proposal: yes, until one verifies) | Owner | Open |
| Q-025 | Auth | How is the Tenhou "ID secret" verified without storing it? (login check via Tenhou? hash compare?) — research needed | Claude | Open |
| Q-030 | Legal | Email C-EGG (support@c-egg.com) for explicit permission? (owner chose to follow rules: no public log links) | Owner | Open |
| Q-031 | Tenhou | Backfill beyond the ~1 year in `list.cgi?old`: yearly `scraw*.zip` currently 404 — where to get history (2009–2025)? Ask Nodocchi's author? | Claude/Owner | Open |
| Q-032 | Tenhou | `mjlog2json.cgi` returns empty body to curl — what headers does it need? (fallback: convert XML ourselves) | Claude | Open |
| Q-033 | Tenhou | Response format of `tenhou.net/mjlog.html` for premium users (owner to supply a paid ID). | Owner | Blocked |
| Q-034 | Tenhou | Should `sca` (個室) / `scf` (大会) results be ingested and shown (filterable), or only ranked `scb`/`scc`? | Owner | Open |
| Q-040 | Data | Confirm tenhou.net/6 JSON as canonical interchange format (D-021). | Owner | Open |
| Q-041 | Data | Store raw logs where? Supabase Storage vs GCS vs R2 (cost/egress). | Claude | Open |
| Q-042 | Stats | Majsoul rank-point and uma tables per room (Gold/Jade/Throne, 3p/4p, Konten) — verify against official sources. | Claude | Open |
| Q-043 | Stats | Ties in placement (Tenhou breaks by seat; Majsoul too) — confirm per platform. | Claude | Open |
| Q-050 | Extension | Extension auth to anyRiichi: Supabase session in extension vs. personal API token? | Claude | Open |
| Q-051 | Extension | Chrome Web Store + Firefox AMO publishing accounts (Chrome $5 one-time) — owner to create. | Owner | Open |
| Q-052 | Extension | Majsoul extension: what is the API for listing a user's paipu, and does using the user's own token in-page breach ToS? (user-side, low risk) | Claude | Open |
| Q-060 | Majsoul | Owner to request an Amae Koromo API key from SAPikachu (for `player_records`). | Owner | Open |
| Q-061 | Majsoul | Which server (CN/JP/EN) for our fetcher account? CN supports password login; EN/JP need token refresh flow. | Claude/Owner | Open |
| Q-062 | Majsoul | Should we ask SAPikachu if AK can share full logs/UUIDs, or collaborate? | Owner | Open |
| Q-070 | Infra | Domain name (anyriichi.com / .net / .app?) and who registers it. | Owner | Open |
| Q-071 | Infra | Supabase plan: Free (500MB, pauses after 1 week idle) vs Pro $25/mo — when to upgrade? Expected Tenhou volume ≈ several M games/yr. | Owner | Open |
| Q-072 | Infra | GCP project/billing account ownership; Vercel team vs personal. | Owner | Open |
| Q-080 | UI | Brand: logo/wordmark, colour palette — mockups use a placeholder 中 tile mark + teal. | Owner | Open |
| Q-081 | UI | Logo usage terms for Mahjong Soul / Tenhou icons. | Claude | Open |
| Q-082 | i18n | Japanese copy: who reviews translations? | Owner | Open |
