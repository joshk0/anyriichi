# User interface

Status: **Proposed — direction reviewed by owner 2026-11-25** · Mockups: [design canvas](https://claude.ai/artifact/7Xck63BrfEFUchViwVWgRv) (private) · [repo mirror](mockups/)

## Screens (mocked)
1. **Player page — desktop**: identity chips (platform logo, rank, pt), filter bar (platform All/MS/Tenhou, 4P/3P, lobby chips, East/South, Ranked/Tournament, period), KPI row, placement distribution per platform, rank-progress chart, stats tabs (Overview, Riichi, Calls & defense, Hand value, Luck, Opponents) with percentile bars, recent-games table (normalized ±pt + raw score, rank-pt delta, promotion/demotion badge & tinted row, log link per visibility).
2. **Player page — phone**: same content stacked; filter chips scroll horizontally; game cards with result, rank-pt delta, promotion badge.
3. **Settings**: one linked account per platform; Tenhou verification paths; log privacy (Private default); sign-in (Google/Email/Discord).
4. **Submit game logs**: unified paste box with platform detection & ranked-only validation; games-missing-a-log list; extension card (MS + Tenhou); Tenhou premium import; detailed-stats coverage.

## Conventions
- Placement colours (colour-blind-safe, differ in lightness): 1st `#1E3A8A`, 2nd `#5B8DEF`, 3rd `#C4C7CF`, 4th `#E8590C`.
- Fonts: Inter (UI), Space Grotesk (numbers/headings), Noto Sans JP (Japanese).
- Accent teal `#0F6E66` — placeholder brand (Q-080).
- Official platform icons for platform identity (D-018, Q-081).
- Every stat has a formula tooltip; metrics added later show "data since".
- Mobile-first: touch targets ≥ 44 px, no horizontal page scroll; wide tables scroll inside a box.

## Not yet mocked
Home/search, leaderboards, combined user page `/u/{handle}`, sign-up/onboarding flow, Japanese locale, game detail page, AI-review export menu, admin/operator console.
