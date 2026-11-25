# Thanks

anyRiichi stands on work done by other people. This file records every project, site and data
source whose information, data, code or ideas we have used — what we took, and where it lives in
this repo. **Add an entry whenever we build on someone else's work** (see CLAUDE.md).

Taking facts (rules, formats, IDs) from a project is not the same as copying its code. Where we
ported logic, the licence is noted; projects without a licence are only read for understanding and
reimplemented.

## Inspiration for the product

| Who | What we took | Where |
|---|---|---|
| **Amae Koromo** (雀魂牌谱屋) by SAPikachu — [amae-koromo.sapk.ch](https://amae-koromo.sapk.ch), [source](https://github.com/SAPikachu/amae-koromo) (MIT) | The whole idea of a player-centric Mahjong Soul stats site; player page layout, stat set, percentile histograms, stable-rank estimate, filters, leaderboards | `docs/research/amae-koromo.md`, `docs/design/product/related-projects.md`, mockups |
| **Nodocchi** (のどっち) — [nodocchi.moe](https://nodocchi.moe) | The Tenhou counterpart: dan/R estimation by replaying games, rank progression charts, promotion/demotion log, streaks, time-of-day analysis | `docs/research/tenhou-and-nodocchi.md`, `docs/design/storage/statistics-model.md` |

## Rules and data

| Who | What we took | Where |
|---|---|---|
| **Tenhou** (天鳳) by C-EGG — [tenhou.net/man](https://tenhou.net/man/) | Official ranked rules: placement points per lobby, last-place losses, promotion thresholds, starting points, R formula; log-ID and rule-code formats; `sc/raw` result files and usage rules | `packages/shared/rules/tenhou.json`, `packages/shared/src/tenhouRule.ts`, `docs/research/tenhou-and-nodocchi.md` |
| **Amae Koromo** — `src/data/types/{level,metadata,gameMode}.ts` @ 78bb145 (MIT) | Mahjong Soul level ids, level max points, last-place penalties, room bonuses, Konten (Celestial) rules and the level-adjustment algorithm (ported to TS and Go) | `packages/shared/rules/majsoul.json`, `packages/shared/src/rules.ts`, `worker/internal/rules` |
| **4n3u/MajsoulData** — [github](https://github.com/4n3u/MajsoulData) | Export of Mahjong Soul's own `desktop/matchmode` table: room mode ids and placement bonuses for every ranked room (used to verify, facts only) | `packages/shared/rules/majsoul.json` (D-049) |
| **Japanese Mahjong Wiki** — [riichi.wiki/Mahjong_Soul](https://riichi.wiki/Mahjong_Soul) | Mahjong Soul Bronze/Silver room bonuses and rank-point worked examples | `packages/shared/rules/majsoul.json` (D-048) |
| **npmahjong** — [rank-up guide](https://npmahjong.com/blog/mahjong-soul-rank-up-guide) | Cross-check of Mahjong Soul room bonuses and per-placement totals | D-048 |
| **RiichiNomi/ronnie** — [github](https://github.com/RiichiNomi/ronnie) | Confirmation that custom-room `mode` values 11–14 are 3-player | research only |
| **vg-mjg/majsoul-api** — [github](https://github.com/vg-mjg/majsoul-api) (MIT) | Reference for talking to Mahjong Soul's API and its protobuf definitions | research only (M2) |

## Formats, converters and review tools

| Who | What we took | Where |
|---|---|---|
| **tenhou.net/6 JSON format** (Tenhou) | Canonical interchange format for full game logs (D-021) | `docs/design/worker/log-conversion.md` |
| **tensoul** (MIT), **Majsoul-to-NAGA** by honvl — [github](https://github.com/honvl/Majsoul-to-NAGA) (MIT) | Reference for converting Mahjong Soul records to tenhou.net/6 | `docs/research/platforms-and-review-tools.md` |
| **Euophrys/tenhou-log** — [github](https://github.com/Euophrys/tenhou-log) | Tenhou client `localStorage` log-list format (for the browser extension) | `docs/design/extension/browser-extension.md` |
| **Benawii's Tenhou log collector** | Windows client log-list format | `docs/research/platforms-and-review-tools.md` |
| **CityLogs** (`rc.honk.li`), cht33's Riichi City parser, EpicOrange's converter, InjusticeJudge | Riichi City log handling (read for understanding; post-v1) | `docs/research/platforms-and-review-tools.md` |
| **Mortal** / [mjai.ekyu.moe](https://mjai.ekyu.moe/), **NAGA** ([naga.dmv.nico](https://naga.dmv.nico/)), **Gokujan** (ex-Bigcoach) | AI-review targets we link out to; input formats they accept | `docs/research/platforms-and-review-tools.md` |

## Built with

Open-source software this repo depends on directly: Next.js and React (MIT), Tailwind CSS (MIT),
Supabase and its CLI (Apache-2.0), PostgreSQL (PostgreSQL licence), pgx (MIT), sqlc (MIT),
Vitest (MIT), TypeScript (Apache-2.0), Go (BSD-3-Clause). Fonts: Inter, Space Grotesk and
Noto Sans JP (SIL OFL) via Google Fonts.
