import { describe, expect, it } from "vitest";
import vectors from "../testdata/log-refs.json";
import { parseLogRef, parseLogRefs } from "./logRef";

describe("parseLogRef (shared vectors)", () => {
  for (const c of vectors.cases) {
    it(c.input, () => {
      expect(parseLogRef(c.input)).toEqual(c.expect);
    });
  }
});

describe("parseLogRefs", () => {
  it("splits pasted input into lines", () => {
    const out = parseLogRefs("2026112412gm-00a9-0000-169b632e\n\n  nonsense  \n");
    expect(out.map((r) => r.line)).toEqual(["2026112412gm-00a9-0000-169b632e", "nonsense"]);
    expect(out[1]!.result).toBeNull();
  });
});
