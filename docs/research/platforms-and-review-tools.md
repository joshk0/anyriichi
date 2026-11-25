# Research: Mahjong Soul API, Riichi City, log conversion, AI review tools, rule differences

> Researched 2026-11-25. **[V]** verified, **[S]** secondary source, **[I]** inferred.

## Mahjong Soul API

- Protobuf RPC over WebSocket; definitions in `liqi.json` (from game client resources); messages wrapped in `lq.Wrapper {name, data}`. `lq.Lobby` has ~391 methods. Proto churns with client updates (MahjongRepository updated 2026-11-16). [V]
- Connection (MahjongRepository/mahjong_soul_api `example.py`) [V]: `GET game.maj-soul.com/1/version.json` → `/1/v{ver}/config.json` (gateway) → `{gateway}/api/clientgate/routes` → `wss://{route.domain}/gateway` → `ReqRequestConnection(route_id)` → login.
- Login:
  - **CN**: `ReqLogin` with `password = HMAC-SHA256(key="lailai", pw)`, `client_version_string="web-{ver}"`. Only CN supports user/password. [V]
  - **EN/JP**: Yostar passport (email code / social) → `uid` + `token` → `oauth2Auth(type=22, code=token, uid)` → `access_token` → `oauth2Check` → `oauth2Login(type=22, tag="en")`. InjusticeJudge uses `wss://engs.mahjongsoul.com:443/gateway` (EN), `gateway-hw.maj-soul.com` (CN). vg-mjg/majsoul-api `passport.ts` (MIT, TS) automates passport login. tensoul: copy `GameMgr.Inst.access_token` from devtools. [V]
  - → EN/JP need a seeded token/cookie with periodic refresh → plan a credential-refresh admin flow. [I]
- RPCs [V]:
  | Method | Req → Res |
  |---|---|
  | `fetchGameRecord` | `{game_uuid, client_version_string}` → `{head: RecordGame, data, data_url}` |
  | `fetchGameRecordList` | own games only |
  | `fetchGameRecordListV2` + `fetchGameRecordsDetailV2` | own games, filters |
  | `fetchGameLiveList` | `{filter_id}` → live games per room |
  | `fetchAccountStatisticInfo` | `{account_id}` → statistic_data[] + detail_data |
  | `fetchAccountInfo` | `{account_id}` → nickname, level… (signature → verification idea) |
  | `searchAccountByPattern` | `{pattern}` |
  | contest: `fetchCustomizedContestGameRecords`, `…GameLiveList` | |
  `fetchGamePlayerList` does **not** exist.
- Record format [V]: `data` (or `data_url`) is a Wrapper around `GameDetailRecords {records[], version, actions[]}`; old logs use `records[]` (`RecordNewRound`, `RecordDealTile`, `RecordDiscardTile`, `RecordChiPengGang`, `RecordAnGangAddGang`, `RecordBaBei`, `RecordHule`, `RecordLiuJu`, `RecordNoTile`), new logs use `actions[]` — parse both. `RecordGame.config.mode.detail_rule` has `init_point`, `fandian`, `shunweima_2..4`, `have_zimosun` → classify ruleset per game.
- Share link `?paipu=<uuid>_a<obfuscated account id>` [S].
- Go libs: constellation39/majsoul gone (404); Pragmatism0220/majsoul (no license, 2023); oscarfzs/majsoulgo (MIT, 2021, stale); Yesterday17/majsoul_api (2022). **Recommendation:** generate Go protos from current `liqi.json` ourselves + thin WS RPC layer. [I]
- **ToS** ([Yostar](https://mahjongsoul.game.yo-star.com/v0.8.1.w/docs/term_of_service.txt) §1.D) bans scraping, bots, harvesting account names, deciphering transmissions, non-official clients, multiple accounts. Amae Koromo operates by tolerance, not permission. Mitigate: throwaway accounts, low rates, no social/payment features, publish nicknames only. [V/I]

## Riichi City

- `zx-api.pesiu.org` = log export for **掌心麻将** (CN version); `/raw?paipuId=`, `/convert?paipuId=&resetInitPoints=`. Sample id `dab6gskj4ak0jap8ge00` → `402 paipuid not exist` (CN-only server). [V/I]
- **CityLogs `rc.honk.li`** [V]: `POST https://rc.honk.li/api/log` with `log_id` → tenhou.net/6 JSON. No auth, no documented limits. Sample worked: 3p East, `ratingc:"PF3"`. `@N` suffix = seat POV.
- Native API [V] (InjusticeJudge `riichicity.py`, Badtaro/RiichiCityTournamentBot): HTTPS POST JSON mimicking Unity client. Domain from `https://d3qgi0t347dz44.cloudfront.net/release/notice/domain_name.ncc` (e.g. `aga.mahjong-jp.net`). Session in a `Cookies` header (JSON: channel, lang, deviceid, sid, uid, region, platform, version). Login: `/users/checkVersion` → `/users/initSession` → `/users/emailLogin {email, passwd(MD5), adjustId}`. Endpoints: `/record/getRoomData {keyValue}` (full log), `/record/readPaiPuList` (records list; isSelf), `/record/readOnlineRoom` (tournament live), `/lobbys/*` (tournament mgmt). No public-lobby live list or profile endpoint found — needs traffic capture.
- Log format: `data.handRecord[]` per kyoku with `handEventRecord[]` events (`eventType` 1 deal, 2 draw, 3 call-wait, 4 action…); hex tile ids (0x01–09 pin, 0x11–19 sou, 0x21–29 man, 0x31–0x91 honors, 0x105/0x115/0x125 red fives).
- Avoid Akagi/Riki (real-time MITM assistants = cheating tools).
- ToS [S]: Steam EULA art. 11 bans bots/external programs; RC announced permanent bans for external tools.

## Conversion & AI review

- **tenhou.net/6 JSON** — `{title, name[4], rule:{disp, aka}, ref, ratingc, lobby, ver, log:[kyoku…]}`; kyoku = `[[idx, honba, sticks], scores[4], dora[], ura[], haipai/draws/discards ×4, result]`; tiles 11–19 m, 21–29 p, 31–39 s, 41–47 honors, 51/52/53 red; `60` tsumogiri; `"r<tile>"` riichi; melds `"c121314"`, `"p191919"`, `m/k/a` kans, `"f44"` kita. **Proposed as anyRiichi's canonical interchange format** — every reviewer accepts it. [V/I]
- honvl/Majsoul-to-NAGA (MIT): Tampermonkey userscript hooking the Majsoul WS, converts to tenhou.net/6, posts one `tenhou.net/5/#json=` URL per round to NAGA's internal `custom_haihu_analyze` with the user's NAGA session + CSRF. 4p only.
- Majsoul → tenhou.net/6: tensoul (MIT, Node, HTTP server `/convert?id=`), tensoul-py.
- **NAGA**: paid (Niconico channel, ~¥550–2,200/mo; custom haihu ~10 NP/round/model), **no public API** → deep link / give user a file; never hold their credentials. [S/I]
- **Mortal** AGPL-3.0 (weights not public); **mjai-reviewer** Apache-2.0 (tenhou.net/6 + seat). **mjai.ekyu.moe** accepts Tenhou/Majsoul URLs, RC log ids, custom tenhou.net/6 JSON; 4p only; free; no API → link out. [V/S]
- **Bigcoach** `review.bigcoach.work` now redirects to **極雀 Gokujan** (gokujan.com); supports Tenhou, Majsoul, RC, ZhangXin, Queji; strong at sanma; paid since 2026-06 (~15 free/month); no API. [V/S]
- RC → tenhou.net/6: cht33 parser (no license), CityLogs, zx-api (CN), EpicOrange/standard-mjlog-converter (no license). **Reimplement, don't copy** unlicensed code.

## Rule differences affecting stats [S]

| | Tenhou | Mahjong Soul | Riichi City |
|---|---|---|---|
| 4p start/return | 25k/30k (oka +20) | 25k/25k (1st needs 30k to end) | 25k/30k |
| 4p uma | +20/+10/−10/−20 | score screen +15/+5/−5/−15; rank pt per room | +30/+10/−10/−30 |
| 3p start/return | 35k/40k | 35k/35k (disputed) | 35k/40k |
| 3p uma | +20/0/−20 | +15/0/−15 | +20/0/−20 |
| 3p tsumo loss | no | yes | ? |
| Red fives | 3 (4p) | 3 (4p), 2 (3p) | 3 (4p), 2 (3p) |
| Triple ron | abortive draw | all pay | double ron allowed |
| Rating | R1500 + dan pt | rank pt by room | dan pt by room (Star/Moon/Sun) |

Implications: store raw score + ruleset id per game; placement-based and per-hand rates are comparable across platforms, points/ratings are not; bucket sanma separately; read Majsoul `detail_rule` per record; flag/exclude event modes; verify per-room rank-point tables manually (riichi.wiki blocked fetch).
