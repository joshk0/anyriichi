import { Suspense } from "react";
import { tenhouRankName } from "@anyriichi/shared";
import { getDictionary, type Dictionary } from "../../dictionaries";
import { getTenhouPlayer, type GameRow } from "@/lib/tenhou-data";
import { PlacementBar } from "@/components/placement-bar";

const placeTint = ["text-place-1", "text-place-2", "text-muted", "text-place-4"];

function room(r: GameRow["games"]["rulesets"], dict: Dictionary): string {
  const players = r.players === 4 ? dict.player.fourP : dict.player.threeP;
  const lobby = dict.lobby[r.lobby as keyof Dictionary["lobby"]] ?? r.lobby;
  const length = r.length === "south" ? dict.player.south : dict.player.east;
  return `${players} ${lobby} ${length}${r.speed !== "normal" ? ` (${dict.player.fast})` : ""}`;
}

const signed = (n: number) => (n > 0 ? `+${n.toFixed(1)}` : n.toFixed(1));

async function Player({ params, dict }: { params: PageProps<"/[lang]/tenhou/[name]">["params"]; dict: Dictionary }) {
  const { lang, name } = await params;
  const decoded = decodeURIComponent(name);
  const data = await getTenhouPlayer(decoded);
  if (!data) {
    return (
      <>
        <h1 className="font-display text-3xl font-bold">{decoded}</h1>
        <p className="text-muted">{dict.player.notFound}</p>
      </>
    );
  }
  // Tenhou tracks 3p and 4p ranks separately; show the latest replayed rank (and pt range) per mode.
  const latestRank = (players: number) => {
    const g = data.recent.find((g) => g.games.rulesets.players === players && g.est_rank_after !== null);
    if (!g || g.est_rank_after === null) return null;
    const pt =
      g.est_pt_min === null || g.est_pt_max === null
        ? ""
        : g.est_pt_min === g.est_pt_max
          ? dict.player.ptExact.replace("{pt}", String(g.est_pt_min))
          : dict.player.ptRange.replace("{min}", String(g.est_pt_min)).replace("{max}", String(g.est_pt_max));
    return { rank: g.est_rank_after, pt };
  };
  const since = new Date(data.player.first_seen_at).toLocaleDateString(lang, { timeZone: "Asia/Tokyo" });
  const placeLabels = (n: number) => ["1", "2", "3", "4"].slice(0, n).map((k) => dict.placement[k as "1"]);
  return (
    <>
      <header className="flex flex-wrap items-baseline gap-3">
        <h1 className="font-display text-3xl font-bold">{data.player.display_name}</h1>
        <span className="rounded-full bg-accent-soft px-2.5 py-1 text-xs font-semibold text-accent-strong">Tenhou</span>
      </header>

      <div className="grid gap-4 md:grid-cols-2">
        {data.modes.map((m) => {
          const avgPlace = m.placements.reduce((s, c, i) => s + c * (i + 1), 0) / m.games;
          const rank = latestRank(m.players);
          return (
            <section key={m.players} className="flex flex-col gap-3 rounded-xl border border-line bg-card p-5">
              <h2 className="flex items-baseline gap-2 font-display text-lg font-semibold">
                {m.players === 4 ? dict.player.fourP : dict.player.threeP}
                {rank && (
                  <span className="text-sm font-normal text-muted">
                    {tenhouRankName(rank.rank, lang === "ja" ? "ja" : "en")}
                    {rank.pt && ` · ${rank.pt}`}
                  </span>
                )}
              </h2>
              <dl className="grid grid-cols-3 gap-3 text-sm">
                <div><dt className="text-muted">{dict.player.games}</dt><dd className="text-xl font-semibold">{m.games}</dd></div>
                <div><dt className="text-muted">{dict.player.avgPlacement}</dt><dd className="text-xl font-semibold">{avgPlace.toFixed(2)}</dd></div>
                <div><dt className="text-muted">{dict.player.avgPt}</dt><dd className="text-xl font-semibold">{signed(m.totalPt / m.games)}</dd></div>
              </dl>
              <PlacementBar rates={m.placements.map((c) => (100 * c) / m.games)} labels={placeLabels(m.players)} />
            </section>
          );
        })}
      </div>

      <section className="flex flex-col gap-3">
        <h2 className="font-display text-xl font-semibold">{dict.player.recent}</h2>
        <div className="overflow-x-auto rounded-xl border border-line bg-card">
          <table className="w-full min-w-[560px] text-sm">
            <thead className="text-left text-muted">
              <tr className="border-b border-line">
                <th className="px-4 py-2 font-medium">{dict.player.date}</th>
                <th className="px-4 py-2 font-medium">{dict.player.room}</th>
                <th className="px-4 py-2 font-medium">{dict.player.place}</th>
                <th className="px-4 py-2 text-right font-medium">{dict.player.score}</th>
                <th className="px-4 py-2 text-right font-medium">{dict.player.pt}</th>
                <th className="px-4 py-2 text-right font-medium">{dict.player.rankPt}</th>
              </tr>
            </thead>
            <tbody>
              {data.recent.map((g, i) => (
                <tr key={i} className={`border-b border-line last:border-0 ${g.rank_event === "promotion" ? "bg-accent-soft" : g.rank_event === "demotion" ? "bg-place-4/10" : ""}`}>
                  <td className="px-4 py-2 whitespace-nowrap">
                    {new Date(g.games.started_at).toLocaleString(lang, { timeZone: "Asia/Tokyo", month: "short", day: "numeric", hour: "2-digit", minute: "2-digit" })}
                  </td>
                  <td className="px-4 py-2 whitespace-nowrap">{room(g.games.rulesets, dict)}</td>
                  <td className={`px-4 py-2 font-semibold ${placeTint[g.placement - 1]}`}>{dict.placement[String(g.placement) as "1"]}</td>
                  <td className="px-4 py-2 text-right tabular-nums">{g.final_score?.toLocaleString(lang) ?? "—"}</td>
                  <td className="px-4 py-2 text-right tabular-nums">{signed(Number(g.result_pt))}</td>
                  <td className="px-4 py-2 text-right tabular-nums whitespace-nowrap">
                    {g.rank_event && (
                      <span className={`mr-2 rounded-full px-2 py-0.5 text-xs font-semibold ${g.rank_event === "promotion" ? "bg-accent text-white" : "bg-place-4 text-white"}`}>
                        {dict.player[g.rank_event]}
                      </span>
                    )}
                    {g.rank_pt_delta == null ? "—" : g.rank_pt_delta > 0 ? `+${g.rank_pt_delta}` : g.rank_pt_delta}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        <p className="text-xs text-muted">{dict.player.rankEstimated}</p>
        <p className="text-xs text-muted">{dict.player.sample.replace("{n}", String(data.sampled)).replace("{since}", since)}</p>
      </section>
    </>
  );
}

export default async function TenhouPlayerPage({ params }: PageProps<"/[lang]/tenhou/[name]">) {
  const dict = await getDictionary();
  return (
    <main className="mx-auto flex max-w-[1200px] flex-col gap-6 px-6 py-10">
      <Suspense fallback={<p className="text-muted">…</p>}>
        <Player params={params} dict={dict} />
      </Suspense>
    </main>
  );
}
