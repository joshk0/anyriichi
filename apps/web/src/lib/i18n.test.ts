import { describe, expect, it } from "vitest";
import { negotiateLocale } from "./i18n";

describe("negotiateLocale", () => {
  it.each([
    [null, "en"],
    ["ja-JP,ja;q=0.9,en;q=0.8", "ja"],
    ["en-US,en;q=0.9,ja;q=0.5", "en"],
    ["fr-FR,ja;q=0.7", "ja"],
    ["fr-FR", "en"],
    ["en;q=0.2,ja;q=0.8", "ja"],
  ])("%s -> %s", (header, want) => {
    expect(negotiateLocale(header)).toBe(want);
  });
});
