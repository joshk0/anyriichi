# Research: Tenhou public data & Nodocchi.moe

> Researched 2026-11-25. **[V]** verified by fetching/reading code, **[I]** inferred/from memory.

## Tenhou raw data — https://tenhou.net/sc/raw/

### Endpoints [V]
- `https://tenhou.net/sc/raw/list.cgi` — last 7 days, JSONP: `list([{file:'sca20261106.log.gz',size:55496}, ...])`. No CORS.
- `https://tenhou.net/sc/raw/list.cgi?old` — older archive (~1 year), year-prefixed: `{file:'2026/scc20260101.html.gz',...}`.
- File URL: `https://tenhou.net/sc/raw/dat/` + file. gzip, `last-modified`, no CORS.
- Recent `scb`/`scc` are **hourly** (`scc2026112412.html.gz`, HH = JST); `sca`/`scf` daily. Archived files are daily for all types.
- Yearly zips `scraw2006.zip`…`scraw2025.zip` are linked but currently **404** (maybe only served after year end / elsewhere) [V/I]. → backfill beyond ~1 year is an open question (Q-031).

### File types [V]
| Prefix | Content | Since | Log links? |
|---|---|---|---|
| `sca` | 個室 (private lobbies L/C) | — | no |
| `scb` | 段位戦 — **all** ranked lobbies (般/上/特/鳳, 3p/4p, 東/南, 速) | — | **no** |
| `scc` | 鳳凰卓 only | 2009-02-01 | yes |
| `scd` | 雀荘戦 | 2010-10-14 | yes [I] |
| `sce` | 技能戦 (closed) + 琥珀卓 | 2013 / 2016 | yes |
| `scf` | 公式大会 / 牌譜公開 | 2015-11-04 | yes |

**Consequence:** for 一般/上級/特上 we get results & placements (scb) but **no log ID**. Full logs for those lobbies only via the player's own client history (extension / manual submit).

### Line formats [V]
```
scb:  12:00 | 17 | 四般東喰赤－ | NaNaNaNa(+48.6) 戸塚安行(+9.0) 玉造(-13.5) NoName(-44.1)
scc:  12:02 | 27 | 四鳳南喰赤－ | <a href="http://tenhou.net/0/?log=2026112412gm-00a9-0000-169b632e">牌譜</a> | 壮心不已(+51.2) HACO(+14.5) ...<br>
sca:  L0001 | 00:00 | 四Ｗ東喰赤－ | tune(+28.3) ...
```
Start time (JST) | duration min | rule | [log] | players **in placement order** with final pt (post uma/oka, thousands; sums to 0). Names HTML-escaped in html files. `NoName` = guest. Rows sorted by **start** time → long games get inserted mid-file later; must re-check by size.

Rule string: 三/四 · level 般上特鳳 (or 若銀琥孔 for 祝儀) · 東/南 · 喰/－ · 赤/－ · 速/光/－ · optional 祝０/祝２/祝５.

### Usage rules ("ログの自動ダウンロードについて") [V]
1. Poll **≤ once / 20 min**; only download files listed in the index.
2. Compare `size`; skip files whose size didn't grow (but keep re-checking recent ones).
3. Download only `.gz` files (no uncompressed 「表示」 views).
4. Always build URLs from `https://tenhou.net/sc/raw/dat/`; follow redirects.
5. **One concurrent session max.** Violators may be banned.

Licensing on same page [V]:
- No products competing with Tenhou; no services that don't need Tenhou play.
- **「不特定多数が天鳳の牌譜をダウンロードするサービスは作成できません」** — no service where the general public downloads Tenhou logs.
- **No redistribution of log files or the log index.**
- Other uses: contact support@c-egg.com.
→ Showing derived stats (as Nodocchi does) is the accepted pattern. **Do not** mirror raw files or bulk-serve mjlogs publicly. Link to tenhou.net for replays. (Q-030: email C-EGG for explicit permission?)

### Full game logs
- `https://tenhou.net/0/log/?{logid}` [V] — mjlog XML (`<mjloggm ver="2.3">`), `text/plain`, **`Access-Control-Allow-Origin: *`**, no auth. Contains `<GO type lobby>`, `<UN n0..n3 dan rate sx>`, rounds, final `owari="score,pt,..."` by seat. Tag spec: Euophrys README (INIT, T/U/V/W draws, D/E/F/G discards, N meld bitfields, REACH, AGARI w/ yaku ids, RYUUKYOKU).
- `https://tenhou.net/5/mjlog2json.cgi?{logid}` — tenhou.net/6 JSON; returned 200 with **empty body** to curl (probably Referer/Origin gated) [V/I]. Plan: fetch XML and convert ourselves.
- Older: `https://tenhou.net/3/mjlog2xml.cgi?{id}` with Referer `tenhou.net/3/` (Euophrys; untested).

### Log ID `2026112412gm-00a9-0000-169b632e` [V]
`YYYYMMDDHH` (JST) · `gm` · 4-hex rule type (= GO type) · lobby (`0000` ranked, `0001` = L0001, 5 digits with leading 1 = C lobby) · 8-hex hash. Older client IDs may have an `x`-prefixed obfuscated 12-hex tail (XOR table in Benawii's `convert_log_id`).

Rule-type bits: 0x01 online · 0x02 **no** aka · 0x04 **no** kuitan · 0x08 南 (else 東) · 0x10 3p · level `((b&0x20)>>4)|((b&0x80)>>7)` → 0x00 般, 0x80 上, 0x20 特, 0xA0 鳳 · 0x40 速 (光 if 0x1000) · 0x200/0x400 祝儀 · 0x800 東天紅 · 0x2000/0x4000 拡張. E.g. `00a9` = 4p 鳳南喰赤, `00b9` 3p 鳳南, `0009` 4p 般南.

`&tw=` = seat (0–3) to view from; web client computes `(4 - oya) % 4`.

## Player's own log list (for the browser extension)

### Web client (`tenhou.net/3/latest.js`) [V]
- `tenhou.net` **localStorage** ring buffer: `lognext` (next index), `log0`…`log39` (JSON strings). Newest-first: `(lognext + 39 - i) % 40`.
- Written at game start: `{"type":169,"lobby":0,"log":"2026...gm-00a9-0000-xxxxxxxx","oya":<dealer rel. to self>,"uname":[self,...]}`; on game end adds `"sc":"score0,pt0,score1,pt1,..."` (relative seats).
- Extension content script on tenhou.net can read this and fetch `/0/log/?id` (CORS open) — or just submit IDs to us.

### Euophrys/tenhou-log (2019) [V]
Reads `log%d` from Firefox `webappsstore.sqlite`, Chrome leveldb, or Selenium; downloads via `mjlog2xml.cgi`. No Windows-client support.

### Windows client [V — Benawii/tenhou-log-collector]
`%USERPROFILE%\AppData\Local\C-EGG\tenhou\130\config.ini`, `[LOG]` section, `n=file={logid}&un0=…&oya=…&type=…&sc=…`, last 40 games. Optional auto-save `.mjlog`.

## Nodocchi.moe

### API [V]
- `GET https://nodocchi.moe/api/listuser.php?name={urlenc}[&start={ts}]` — Cloudflare, JSON, no CORS. All games since ~2009.
```json
{"name":"HACO","recent":1795614000,"rate":{"4":2131},"rseq":[[1600176360,1795614000]],
 "list":[{"starttime":1795614000,"during":13,"sctype":"c","playernum":4,"playerlevel":3,
   "playlength":2,"kuitanari":1,"akaari":1,"speed":0,"player1":"...","player1ptr":"56.0", ...,
   "url":"http://tenhou.net/0/?log=...","tw":"228","rate":2129}]}
```
`sctype` = source file a–f; `playerlevel` 0般 1上 2特 3鳳; `playlength` 0技能 1東 2南; players in placement order; optional `lobby`, `shuugi`, `seq` (anchor for dan estimate), per-game `rate`. `tw` packs 2 bits/seat = placement−1.
- Per-game R `rate` exists for 特上/鳳凰 games (~99%) — likely from Tenhou's live spectate feed [I]; absent for lower lobbies.
- `phoenix_status.php?all=1&username=` — per-hand stats from 鳳凰 mjlogs (agariC, houjuuC, riichC, fuuroC, tsumoV, damaV, tobiZ, agariVT, houjuuVT, agariVJ, ura/ippatsu/aka, yaku rates, all-last placement change, …).
- `/s/ugr/{name}.js` dan ranking; `/s/wg.js` live games; others seen in JS.

### Features (from `min/base.js`) [V]
- Header: estimated dan+pt (推定段位) & highest; estimated R & highest; dan ranking; account lifetime; idle days; recent-results strip.
- Results summary: placement counts/%; 平順 avg placement; 平得 avg pt; 通得 total pt; max/min score; 祝儀; point histogram incl. reconstructed raw score (4p uma `[40,10,-10,-20]`−30 / 3p `[25,0,-20]`−40).
- **Charts:** dan/pt and R progression as **candlestick (OHLC)**, split by N games / dan change; linear-rank option; "virtual grade" beyond 天鳳位; new-high markers.
- Analyses: stable dan (games to promotion/demotion); streaks; placement conditioned on gap since previous game (1h/3h/8h/12h/24h); weekday×hour heatmap; frequent opponents & H2H; 点差; grade-change log.
- Game list: #, start, duration, lobby, rule, log link, placement, dan & pt change, R, virtual dan, players with pts; options: colour by placement, JST toggle, link variants, **CSV export**, bulk mjlog download.
- Filters (URL hash `#!&name=...`): date range, type (個室/大会, 段位戦, 雀荘/技能), 3p/4p, level, 東/南, 喰/赤, speed, score range, point-gap, has-log, promotion games only, lobby regex, last-N, with co-players.
- Other pages: grade changes, dan ranking, lobby lookup, leaderboards (特上/鳳凰), most-searched, phoenix stats ranking, live spectate with notify-on-player, dark theme.

### Dan/R estimation (`THR.ptcfg`) [V]
- Dan start/promotion pt: 初段 200/400, 二段 400/800, 三段 600/1200, 四段 800/1600, 五段 1000/2000, 六段 1200/2400, 七段 1400/2800, 八段 1600/3200, 九段 1800/3600, 十段 2000/4000, 天鳳位.
- Kyu thresholds: 新人–7級 20, 6級 40, 5級 60, 4級 80, 3級–1級 100. Demotion only from 初段.
- Last-place loss: 0 to 3級, 2級 −10, 1級 −20, 初段 −30, then −10 per dan to 十段 −120.
- Gains 4p 1st/2nd (3p 1st): 般 +20/+10 (+30), 上 +40/+10 (+50), 特 +50/+20 (+70), 鳳 +60/+30 (+90). 南 ×1.5.
- History: pre-2008-08-31 南 multiplier 1; pre-2017-10-24 kyu thresholds 30/30/30/60/60/60/90 and 一般 4p +30/+0.
- Estimation: replay every ranked (b/c) game from first game; re-anchor at known `seq`/`rseq` or after >180 days idle.
- R (Tenhou, from memory [I]): `ΔR = max(1 − 0.002·games, 0.2) × (bonus + (tableAvgR − R)/40)`, start 1500, bonus 4p `[+30,+10,−10,−30]`, 3p `[+30,0,−30]`.
- Tobi only from 鳳凰 mjlogs (or infer from negative reconstructed raw score [I]).
