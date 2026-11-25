export const locales = ["en", "ja"] as const;
export type Locale = (typeof locales)[number];
export const defaultLocale: Locale = "en";

export const hasLocale = (l: string): l is Locale => (locales as readonly string[]).includes(l);

/** Pick a supported locale from an Accept-Language header (q-values respected, region ignored). */
export function negotiateLocale(acceptLanguage: string | null): Locale {
  if (!acceptLanguage) return defaultLocale;
  const prefs = acceptLanguage
    .split(",")
    .map((part) => {
      const [tag, ...params] = part.trim().split(";");
      const q = params.map((p) => p.trim()).find((p) => p.startsWith("q="));
      return { lang: (tag ?? "").toLowerCase().split("-")[0] ?? "", q: q ? Number(q.slice(2)) : 1 };
    })
    .filter((p) => p.lang && !Number.isNaN(p.q))
    .sort((a, b) => b.q - a.q);
  return prefs.map((p) => p.lang).find(hasLocale) ?? defaultLocale;
}
