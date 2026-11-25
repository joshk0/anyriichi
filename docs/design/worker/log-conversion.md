# Log parsing, conversion & AI-review export

Status: **Proposed** (canonical format pending Q-040)

## Pipeline
`raw platform log → platform parser → internal event model → (a) per-hand stats, (b) tenhou.net/6 JSON export`

- **Internal event model** (Go structs, mirrored in TS types): game header (players, ruleset), rounds with deal, draws, discards (tsumogiri flag, riichi), calls, kans, kita, wins (han/fu/yaku/points/payments), draws (type, tenpai), score deltas.
- **Parsers:** Tenhou mjlog XML (M1), Majsoul protobuf records (M2), Riichi City JSON (later).
- **Export:** tenhou.net/6 JSON per game (and per-round split for NAGA's custom form). Reimplement converters ourselves; consult MIT tensoul / Majsoul-to-NAGA for edge cases; do not copy unlicensed code.
- **Golden tests:** fixtures of real logs (from owner's `.env`-referenced IDs; fixtures themselves stored privately, not committed if ToS forbids) with expected stats.

## AI review (M4)
| Reviewer | Integration |
|---|---|
| Mortal / mjai.ekyu.moe | deep link with Tenhou/Majsoul URL; or download tenhou.net/6 JSON to paste |
| NAGA | download tenhou.net/6 JSON (per-round split); NAGA is paid, no API — never hold user NAGA credentials |
| Gokujan (ex-Bigcoach) | deep link / file |

Export respects log visibility (only users allowed to view a log may export it).
