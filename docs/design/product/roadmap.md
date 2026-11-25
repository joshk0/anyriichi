# Roadmap

Status: **Proposed** · Based on D-005 (Tenhou-first), D-001/D-006, D-017.

| Milestone | Goal | Key contents |
|---|---|---|
| **M0 Foundations** | Repo, infra & skeletons | Monorepo layout, Supabase project, Vercel app, GCE worker VM, CI, DB migrations, i18n scaffolding, design-system tokens |
| **M1 Tenhou MVP** | Public Tenhou player pages | `scb`/`scc` ingestion (all public games, ~1 yr backfill), players & search, player page (placement stats, normalized ±pt, dan/pt estimate with promotions/demotions, rank chart, game list), sign-up (Google/Email/Discord), link Tenhou account, log submission (paste) + per-hand stats from mjlogs, log privacy, missing-logs list, EN/JA |
| **M2 Mahjong Soul** | Second platform | AK integration (stats + records with key), own `fetchGameRecord` fetcher for registered users, Majsoul parser → per-hand stats, rank-pt deltas, link & verify MS account, cross-platform "All" view |
| **M3 Extension & import** | Effortless log collection | Browser extension (Chrome/Firefox) for Tenhou + Majsoul, Tenhou premium `mjlog.html` import, Tenhou verification flows (25-log, ID secret, review) |
| **M4 Analysis & community** | Depth | AI-review export (tenhou.net/6 file, NAGA/Mortal/Gokujan deep links), percentile histograms, leaderboards, opponents/H2H, CSV export |
| **Later** | | Riichi City, leagues/clubs, live games, notifications, LINE/Apple login, 2FA, auth-method linking |
