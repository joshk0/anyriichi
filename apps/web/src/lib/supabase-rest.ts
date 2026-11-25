import "server-only";

/**
 * Minimal read-only PostgREST client for public tables (RLS: public read). Server-side only.
 * Uses the anon key; never the service role.
 */
export async function rest<T>(path: string, init?: { count?: boolean }): Promise<{ data: T; count: number | null }> {
  const url = process.env.NEXT_PUBLIC_SUPABASE_URL;
  const key = process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY;
  if (!url || !key) throw new Error("Supabase env vars are not set");
  const res = await fetch(`${url}/rest/v1/${path}`, {
    headers: {
      apikey: key,
      Authorization: `Bearer ${key}`,
      ...(init?.count ? { Prefer: "count=exact" } : {}),
    },
  });
  if (!res.ok) throw new Error(`Supabase ${res.status}: ${await res.text()}`);
  const range = res.headers.get("content-range");
  const count = range && range.includes("/") ? Number(range.split("/")[1]) : null;
  return { data: (await res.json()) as T, count: Number.isFinite(count) ? count : null };
}
