# Mahjong Soul ingestion (M2)

Status: **Proposed** (strategy accepted: D-006) · Research: [/docs/research/amae-koromo.md](../../research/amae-koromo.md), [platforms](../../research/platforms-and-review-tools.md)

## Sources
1. **Amae Koromo API** (`https://{1,2,4,5}.data.amae-koromo.com/api/v2/pl4/` and `pl3`):
   - `player_stats`, `player_extended_stats` (open) → cross-check / bootstrap aggregate stats.
   - `player_records` (needs key from SAPikachu, Q-060) → list of the registered player's games (uuid, mode, players, scores, `gradingScore`).
   - Only for players claimed on anyRiichi (D-001). Respect their mirrors; cache aggressively.
2. **Own record fetcher** (Go):
   - Dedicated throwaway account (server choice Q-061; CN supports password login, EN/JP need token refresh).
   - WebSocket protobuf client generated from current `liqi.json` (track client version; regenerate on update).
   - `fetchGameRecord(uuid)` only for games listed for registered players or submitted by users. Low rate.
   - Parse `records[]`/`actions[]` → hands/hand_player_stats; read `detail_rule` for ruleset classification; convert to tenhou.net/6 JSON.
3. **Extension** (M3): user's own token lists their full paipu history → submissions.

## Risks
Majsoul ToS forbids bots/scraping; mitigate with low volume (registered users only), dedicated account, no social/payment features, kill switch. Protocol churn → alert when login/decoding fails.

## Rooms
Ranked modes: Bronze/Silver (not in AK) + Gold 8/9, Jade 11/12, Throne 15/16 (4p East/South); 3p Gold 21/22, Jade 23/24, Throne 25/26. Friendly/contest modes rejected for submissions (D-012).
