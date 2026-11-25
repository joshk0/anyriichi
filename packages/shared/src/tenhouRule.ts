/**
 * Decode the 4-hex Tenhou rule code (the `GO type` value) found in log IDs, e.g. `00a9`.
 * Bit meanings from tenhou.net client JS — see docs/research/tenhou-and-nodocchi.md#log-id.
 */
export type TenhouLevel = "ippan" | "joukyuu" | "tokujou" | "houou";

export interface TenhouRule {
  players: 3 | 4;
  length: "east" | "south";
  level: TenhouLevel;
  speed: "normal" | "fast" | "hikari";
  aka: boolean;
  kuitan: boolean;
  shuugi: boolean;
}

const LEVELS: readonly TenhouLevel[] = ["ippan", "joukyuu", "tokujou", "houou"];

export function decodeTenhouRule(code: string): TenhouRule {
  const b = parseInt(code, 16);
  if (!/^[0-9a-f]{4}$/i.test(code) || Number.isNaN(b)) {
    throw new Error(`invalid tenhou rule code: ${code}`);
  }
  const level = ((b & 0x20) >> 4) | ((b & 0x80) >> 7);
  return {
    players: b & 0x10 ? 3 : 4,
    length: b & 0x08 ? "south" : "east",
    level: LEVELS[level]!,
    speed: b & 0x40 ? (b & 0x1000 ? "hikari" : "fast") : "normal",
    aka: (b & 0x02) === 0,
    kuitan: (b & 0x04) === 0,
    shuugi: (b & 0x600) !== 0,
  };
}
