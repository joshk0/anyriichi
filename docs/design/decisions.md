# Decision log

Append-only. Each entry: ID, date, decision, rationale, source. Superseding decisions reference the old ID.

| ID | Date | Decision | Source |
|---|---|---|---|
| D-001 | 2026-11-25 | **Data scope is hybrid.** Ingest *all* public Tenhou games. For Mahjong Soul, use Amae Koromo (AK) for registered users, plus our own fetcher only for registered users' game records (no live-room crawling). | Owner, kickoff Q&A |
| D-002 | 2026-11-25 | **Frontend: Next.js (App Router, React, TypeScript)**, Tailwind + shadcn/ui. | Owner |
| D-003 | 2026-11-25 | **Auth + DB: Supabase** (Supabase Auth + Postgres). | Owner |
| D-004 | 2026-11-25 | **Web hosting: Vercel.** Go worker on a **GCE e2-micro (Always Free)** VM. Hobby tier is non-commercial — revisit (Pro $20/mo) if we monetize. | Owner |
| D-005 | 2026-11-25 | **MVP = Tenhou-first.** Mahjong Soul follows in M2; Riichi City after v1. | Owner |
| D-006 | 2026-11-25 | **Majsoul strategy: AK + own fetcher.** Owner asks SAPikachu for an AK API key; our worker uses a dedicated throwaway Majsoul account to call `fetchGameRecord` for registered users' games only. | Owner |
| D-007 | 2026-11-25 | **Tenhou ToS posture: log links are not publicly revealed** (like AK masking). Users' log visibility defaults to **Private**. Showing a log link to a registered user whose *claimed* player is in that game is acceptable. | Owner |
| D-008 | 2026-11-25 | **Log visibility setting** per user: Private (default) / Players at the table / Public. Applies to logs the user submits. | Owner, canvas comment |
| D-009 | 2026-11-25 | **One linked account per platform** per user (no alts) — alts may violate platform ToS. Idea for later: private secondary accounts. | Owner, canvas comment |
| D-010 | 2026-11-25 | **V1 sign-in: Google, Email+password, Discord.** No 2FA in v1. No switching/merging auth methods in v1. *(Supersedes the 2FA requirement in the original spec for v1.)* | Owner, canvas comment |
| D-011 | 2026-11-25 | **Tenhou account verification**: verified when the same user submits **25 ranked logs** containing that player (proposed tightening: logs played after claim opened, spread over ≥ 5 days); or by **ID secret** (instant, optional); or **manual review** to transfer a claimed account. Auto-verified links are revocable by stronger proof. | Owner, canvas comment |
| D-012 | 2026-11-25 | **Only ranked-room logs** may be submitted by users (Tenhou 段位戦 lobbies; Majsoul rank match rooms). Tournament stats from public feeds may still be stored. | Owner, canvas comment |
| D-013 | 2026-11-25 | **Single log-submission page** for all platforms with automatic platform detection from the log ID/URL. | Owner, canvas comment |
| D-014 | 2026-11-25 | **Browser extension v1 supports Tenhou and Mahjong Soul** (MS: use the session token to bulk-list the user's paipu; Tenhou: read localStorage `log0..39`, auto-submit finished games). | Owner, canvas comment |
| D-015 | 2026-11-25 | **Tenhou premium import**: support importing a user's history via `tenhou.net/mjlog.html` by Tenhou ID (owner will provide a paid ID to study the response). | Owner, canvas comment |
| D-016 | 2026-11-25 | **Game results normalized to Tenhou-style ±pt** across platforms using each room's uma/oka; also show **rank-point delta per game**, and highlight **promotions/demotions**. | Owner, canvas comment |
| D-017 | 2026-11-25 | **Riichi City deferred** out of v1 designs (keep layouts extensible for a third platform). | Owner, canvas comment |
| D-018 | 2026-11-25 | **Use official platform logos** in the UI from the start (currently app icons; confirm brand-usage terms before launch). | Owner, canvas comment |
| D-019 | 2026-11-25 | **"Games missing a log" list** so users can find and submit logs for Tenhou games without one. | Owner, canvas comment |
| D-020 | 2026-11-25 | **Mockups live on a Claude design canvas and are mirrored in `docs/design/web/mockups/`** for account portability. | Owner |
| D-021 | 2026-11-25 | **Canonical full-log interchange format = tenhou.net/6 JSON** (all AI reviewers accept it); raw platform logs are kept privately for re-processing. | Proposed by Claude — *needs owner confirmation (Q-040)* |
