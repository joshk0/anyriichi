/**
 * Detect game-log references in user input (D-013, D-040).
 * - Tenhou: match the log ID anywhere; the surrounding URL is ignored.
 * - Mahjong Soul: require a full share URL on one of the three official hosts.
 * Patterns avoid lookaround so the Go (RE2) implementation can use identical regexes.
 */
import { decodeTenhouRule, type TenhouRule } from "./tenhouRule";

export const TENHOU_LOG_ID_RE =
  /(?:^|[^0-9A-Za-z])(\d{10}gm-([0-9a-f]{4})-(\d{4,5})-x?[0-9a-f]{8,16})(?:[^0-9A-Za-z]|$)/;
const TENHOU_SEAT_RE = /[?&]tw=([0-3])(?:[^0-9]|$)/;

export const MAJSOUL_URL_RE =
  /https?:\/\/(mahjongsoul\.game\.yo-star\.com|game\.mahjongsoul\.com|game\.maj-soul\.com)\/[^\s?#]*\?(?:[^\s#]*&)?paipu=(\d{6}-[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})(?:_a(\d+))?(?:[^0-9A-Za-z-]|$)/;
const MAJSOUL_ANY_PAIPU_RE =
  /https?:\/\/(?:mahjongsoul\.game\.yo-star\.com|game\.mahjongsoul\.com|game\.maj-soul\.com)\/[^\s?#]*\?(?:[^\s#]*&)?paipu=/;

export type MajsoulServer = "en" | "jp" | "cn";
const MAJSOUL_HOSTS: Record<string, MajsoulServer> = {
  "mahjongsoul.game.yo-star.com": "en",
  "game.mahjongsoul.com": "jp",
  "game.maj-soul.com": "cn",
};

export interface TenhouLogRef {
  platform: "tenhou";
  id: string;
  seat: number | null;
  lobby: string;
  ruleCode: string;
  rule: TenhouRule;
  /** Ranked (段位戦) lobby. Only ranked logs may be submitted by users (D-012). */
  ranked: boolean;
}

export interface MajsoulLogRef {
  platform: "majsoul";
  server: MajsoulServer;
  uuid: string;
  /** Obfuscated account/seat suffix after `_a`, if present. */
  account: string | null;
}

export type LogRefError = { error: "majsoul_anonymous" };
export type LogRef = TenhouLogRef | MajsoulLogRef;

/** Parse a single line/token of user input. Returns null if it is not a recognised log reference. */
export function parseLogRef(input: string): LogRef | LogRefError | null {
  const t = TENHOU_LOG_ID_RE.exec(input);
  if (t) {
    const [, id, ruleCode, lobby] = t as unknown as [string, string, string, string];
    const seat = TENHOU_SEAT_RE.exec(input);
    return {
      platform: "tenhou",
      id,
      seat: seat ? Number(seat[1]) : null,
      lobby,
      ruleCode,
      rule: decodeTenhouRule(ruleCode),
      ranked: lobby === "0000",
    };
  }
  const m = MAJSOUL_URL_RE.exec(input);
  if (m) {
    const [, host, uuid, account] = m as unknown as [string, string, string, string | undefined];
    return { platform: "majsoul", server: MAJSOUL_HOSTS[host]!, uuid, account: account ?? null };
  }
  if (MAJSOUL_ANY_PAIPU_RE.test(input)) return { error: "majsoul_anonymous" };
  return null;
}

/** Parse multi-line pasted input: one result per non-empty line. */
export function parseLogRefs(text: string): Array<{ line: string; result: LogRef | LogRefError | null }> {
  return text
    .split(/\r?\n/)
    .map((l) => l.trim())
    .filter(Boolean)
    .map((line) => ({ line, result: parseLogRef(line) }));
}
