import Link from "next/link";
import { getDictionary } from "@/app/[lang]/dictionaries";

export async function SiteHeader({ lang }: { lang: string }) {
  const dict = await getDictionary();
  const other = lang === "ja" ? "en" : "ja";
  return (
    <header className="border-b border-line bg-card">
      <div className="mx-auto flex max-w-[1200px] flex-wrap items-center gap-5 px-6 py-3.5">
        <Link href={`/${lang}`} className="flex items-center gap-2.5 text-ink no-underline">
          <span
            aria-hidden="true"
            className="flex h-[38px] w-[30px] items-center justify-center rounded-[5px] border-2 border-ink font-bold text-tile"
          >
            中
          </span>
          <span className="font-display text-[22px] font-bold tracking-tight">anyRiichi</span>
        </Link>
        <nav className="ml-auto flex items-center gap-4 text-sm font-medium">
          <span className="text-muted">{dict.nav.leaderboards}</span>
          <span className="text-muted">{dict.nav.submit}</span>
          <Link
            href={`/${other}`}
            aria-label={dict.nav.language}
            className="rounded-md border border-line-strong px-2.5 py-1.5 text-ink no-underline"
          >
            {other === "ja" ? "日本語" : "English"}
          </Link>
        </nav>
      </div>
    </header>
  );
}
