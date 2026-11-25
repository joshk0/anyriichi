# Open questions

Resolve a question when it becomes relevant for implementation; move the answer into `decisions.md` and mark it
**Resolved → D-xxx** here (keep the row for history). Owner = who must answer.

| ID | Area | Question | Owner | Status |
|---|---|---|---|---|
| Q-010 | Privacy | Can a player **hide their profile** (opt-out) even though the data is public on the source site? Takedown process? | Owner | **Resolved → D-030** |
| Q-011 | Product | Are leagues/clubs (persona P5) in scope for v1? (assumed no) | Owner | **Resolved → D-031** |
| Q-012 | Privacy | Log visibility: account-wide only, or **per-log override** (e.g. make one game public)? | Owner | **Resolved → D-029** |
| Q-013 | Privacy | When table-mates choose different visibility, does the **strictest** setting among claimed participants win, or each uploader's choice? (Tenhou logs show all 4 hands.) | Owner | **Resolved → D-029** |
| Q-020 | Auth | LINE login for JP users (custom OIDC)? (post-v1) | Owner | Deferred |
| Q-021 | Auth | Apple login (needs $99/yr Apple dev account)? (post-v1) | Owner | Deferred |
| Q-022 | Auth | Tenhou verification tightening (post-claim logs, ≥5 days)? | Owner | **Resolved → D-022 (no conditions)** |
| Q-023 | Auth | Majsoul verification: extension-based (token proves account) vs. signature-nonce via `fetchAccountInfo`? Both? | Owner | Open |
| Q-024 | Auth | Can two users hold **unverified** claims on the same player? (proposal: yes, until one verifies) | Owner | Open |
| Q-025 | Auth | How is the Tenhou ID secret verified? Implied by D-022: we use the login info once to prove ownership, then delete it. Mechanism (log in to Tenhou? read account info?) — research needed | Claude | Open |
| Q-030 | Legal | Email C-EGG (support@c-egg.com) for explicit permission? (owner chose to follow rules: no public log links) | Owner | **Resolved → D-032** |
| Q-031 | Tenhou | Backfill beyond the ~1 year in `list.cgi?old`: yearly `scraw*.zip` currently 404 — where to get history (2009–2025)? Ask Nodocchi's author? | Claude/Owner | Open |
| Q-032 | Tenhou | `mjlog2json.cgi` returns empty body to curl — what headers does it need? (fallback: convert XML ourselves) | Claude | Open |
| Q-033 | Tenhou | Response format of `tenhou.net/mjlog.html` for premium users (owner setting up paid account, D-038). | Owner | In progress |
| Q-034 | Tenhou | Should `sca` (個室) / `scf` (大会) results be ingested and shown (filterable), or only ranked `scb`/`scc`? | Owner | **Resolved → D-033** |
| Q-040 | Data | Confirm tenhou.net/6 JSON as canonical interchange format (D-021). | Owner | **Resolved → D-021** |
| Q-041 | Data | Store raw logs where? Supabase Storage vs GCS vs R2 (cost/egress). | Claude | **Resolved → D-039** (Supabase Storage) |
| Q-042 | Stats | Majsoul rank-point and uma tables per room (Gold/Jade/Throne, 3p/4p, Konten) — verify against official sources. | Claude | Resolved → D-045 |
| Q-043 | Stats | Ties in placement (Tenhou breaks by seat; Majsoul too) — confirm per platform. | Claude | Resolved → statistics-model §1 (use platform placement) |
| Q-050 | Extension | Extension auth to anyRiichi: Supabase session in extension vs. personal API token? | Claude | Open |
| Q-051 | Extension | Chrome Web Store + Firefox AMO publishing accounts (Chrome $5 one-time) — owner to create. | Owner | **Resolved → D-038** |
| Q-052 | Extension | Majsoul extension: what is the API for listing a user's paipu, and does using the user's own token in-page breach ToS? (user-side, low risk) | Claude | Open |
| Q-060 | Majsoul | Owner to request an Amae Koromo API key from SAPikachu (for `player_records`). | Owner | **Resolved → D-034** |
| Q-061 | Majsoul | Which server (CN/JP/EN) for our fetcher account? CN supports password login; EN/JP need token refresh flow. | Claude/Owner | Open |
| Q-062 | Majsoul | Should we ask SAPikachu if AK can share full logs/UUIDs, or collaborate? | Owner | **Resolved → D-034** |
| Q-070 | Infra | Domain name (anyriichi.com / .net / .app?) and who registers it. | Owner | **Resolved → D-035** |
| Q-071 | Infra | DB choice & cost: owner asked for a GCP price-out incl. free tiers and overage, and to consider Cloud Firestore (no idle cost, possible secure browser-direct reads) and similar serverless DBs, vs Supabase Free/Pro. Analysis in progress → docs/research/db-cost-analysis.md | Claude → Owner | **Resolved → D-039** |
| Q-072 | Infra | GCP project/billing account ownership; Vercel team vs personal. | Owner | **Resolved → D-035** |
| Q-080 | UI | Brand: logo/wordmark, colour palette — mockups use a placeholder 中 tile mark + teal. | Owner | **Resolved → D-036** |
| Q-081 | UI | Logo usage terms for Mahjong Soul / Tenhou icons. | Claude | Open |
| Q-082 | i18n | Japanese copy: who reviews translations? | Owner | **Resolved → D-037** |
| Q-026 | Auth | Premium import with a Tenhou ID that belongs to a *different* player than the user's already-verified Tenhou claim: block, or ask before switching? | Owner | **Resolved → D-028** |
| Q-083 | UI | Rank-progress chart with several platforms: platform picker (chip per linked platform, default most-played; proposed) vs overlay; optional "Compare" view plotting percentile within each platform? | Owner | **Resolved → D-027** |
