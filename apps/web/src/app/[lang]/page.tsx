import { lang } from "next/root-params";
import { getDictionary } from "./dictionaries";
import { PlacementBar } from "@/components/placement-bar";

export default async function Home() {
  const dict = await getDictionary();
  const locale = await lang();
  return (
    <main className="mx-auto flex max-w-[1200px] flex-col gap-8 px-6 py-12">
      <section className="flex flex-col gap-3">
        <h1 className="font-display text-4xl font-bold tracking-tight sm:text-5xl">{dict.home.headline}</h1>
        <p className="max-w-2xl text-lg text-muted">{dict.home.sub}</p>
      </section>
      <form role="search" action={`/${locale}/search`} className="flex max-w-xl flex-col gap-2">
        <label htmlFor="q" className="text-sm font-semibold">{dict.home.searchLabel}</label>
        <div className="flex gap-2">
          <input id="q" name="q" type="search" required placeholder={dict.home.searchPlaceholder}
            className="h-11 flex-1 rounded-lg border border-line-strong bg-card px-3.5 text-sm" />
          <button type="submit" className="h-11 rounded-lg bg-accent px-4 text-sm font-semibold text-white hover:bg-accent-strong">
            {dict.home.searchSubmit}
          </button>
        </div>
        <p className="text-sm text-muted">{dict.home.comingSoon}</p>
      </form>
      <section className="max-w-xl rounded-xl border border-line bg-card p-5">
        <h2 className="mb-3 font-display text-lg font-semibold">{dict.home.placementDemo}</h2>
        <PlacementBar rates={[27.6, 25.3, 27.0, 20.1]} labels={["1", "2", "3", "4"].map((k) => dict.placement[k as "1"])} />
      </section>
    </main>
  );
}
