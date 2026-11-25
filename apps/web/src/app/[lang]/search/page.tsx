import { Suspense } from "react";
import Link from "next/link";
import { getDictionary, type Dictionary } from "../dictionaries";
import { searchTenhouPlayers } from "@/lib/tenhou-data";

async function Results({ searchParams, dict, locale }: {
  searchParams: PageProps<"/[lang]/search">["searchParams"];
  dict: Dictionary;
  locale: string;
}) {
  const raw = (await searchParams).q;
  const q = (Array.isArray(raw) ? raw[0] : raw)?.trim() ?? "";
  if (!q) return <p className="text-muted">{dict.search.prompt}</p>;
  const players = await searchTenhouPlayers(q);
  if (players.length === 0) return <p className="text-muted">{dict.search.none.replace("{q}", q)}</p>;
  return (
    <ul className="flex flex-col divide-y divide-line rounded-xl border border-line bg-card">
      {players.map((p) => (
        <li key={p.id}>
          <Link href={`/${locale}/tenhou/${encodeURIComponent(p.display_name)}`}
            className="flex items-center justify-between gap-4 px-4 py-3 hover:bg-ground">
            <span className="font-semibold">{p.display_name}</span>
            <span className="text-sm text-muted">
              {dict.search.lastSeen} {new Date(p.last_seen_at).toLocaleString(locale, { timeZone: "Asia/Tokyo" })}
            </span>
          </Link>
        </li>
      ))}
    </ul>
  );
}

export default async function SearchPage({ params, searchParams }: PageProps<"/[lang]/search">) {
  const dict = await getDictionary();
  const { lang } = await params;
  return (
    <main className="mx-auto flex max-w-[1200px] flex-col gap-6 px-6 py-10">
      <h1 className="font-display text-3xl font-bold">{dict.search.title}</h1>
      <Suspense fallback={<p className="text-muted">…</p>}>
        <Results searchParams={searchParams} dict={dict} locale={lang} />
      </Suspense>
    </main>
  );
}
