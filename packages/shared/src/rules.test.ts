import { describe, expect, it } from "vitest";
import v from "../testdata/rules.json";
import {
  applyMajsoulGame,
  applyTenhouGame,
  majsoulLevelName,
  resultPt,
  tenhouLogSupported,
  tenhouRankName,
  type Platform,
  type Players,
  type Length,
} from "./rules";
import type { TenhouLevel } from "./tenhouRule";

describe("resultPt (shared vectors)", () => {
  for (const c of v.resultPt) {
    it(`${c.platform} ${c.players}p ${c.score} #${c.placement}`, () => {
      expect(resultPt(c.platform as Platform, c.players as Players, c.score, c.placement)).toBe(c.expect);
    });
  }
  it("rejects bad placements", () => {
    expect(() => resultPt("tenhou", 3, 30000, 4)).toThrow();
  });
});

describe("applyTenhouGame (shared vectors)", () => {
  for (const c of v.tenhouGames) {
    it(c.name, () => {
      const r = applyTenhouGame(
        { rank: c.rank, pt: c.pt },
        { players: c.players as Players, level: c.level as TenhouLevel, length: c.length as Length, placement: c.placement },
      );
      expect({ delta: r.delta, rank: r.after.rank, pt: r.after.pt, event: r.event }).toEqual(c.expect);
    });
  }
});

describe("applyMajsoulGame (shared vectors)", () => {
  for (const c of v.majsoulGames) {
    it(c.name, () => {
      const r = applyMajsoulGame({ levelId: c.levelId, points: c.points }, { mode: c.mode, score: c.score, placement: c.placement });
      expect({ delta: r.delta, levelId: r.after.levelId, points: r.after.points, event: r.event }).toEqual(c.expect);
    });
  }
});

describe("tenhouLogSupported (shared vectors)", () => {
  for (const c of v.tenhouSupported) {
    it(c.logId, () => expect(tenhouLogSupported(c.logId)).toBe(c.expect));
  }
});

describe("names", () => {
  it("formats ranks", () => {
    expect(tenhouRankName(16, "ja")).toBe("七段");
    expect(tenhouRankName(20)).toBe("Tenhoui");
    expect(majsoulLevelName(10402)).toBe("Master 2");
    expect(majsoulLevelName(20703, "ja")).toBe("魂天 3");
  });
});
