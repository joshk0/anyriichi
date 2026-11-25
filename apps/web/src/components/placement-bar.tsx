const segment = ["bg-place-1 text-white", "bg-place-2 text-ink", "bg-place-3 text-ink", "bg-place-4 text-ink"];

/** Horizontal stacked bar of placement rates (percentages summing to ~100). */
export function PlacementBar({ rates, labels }: { rates: number[]; labels: string[] }) {
  return (
    <div className="flex flex-col gap-1.5">
      <div className="flex h-[26px] overflow-hidden rounded-md text-xs font-semibold" role="img"
        aria-label={rates.map((r, i) => `${labels[i]} ${r.toFixed(1)}%`).join(", ")}>
        {rates.map((r, i) => (
          <div key={i} className={`flex items-center justify-center ${segment[i]}`} style={{ width: `${r}%` }}>
            {r.toFixed(1)}
          </div>
        ))}
      </div>
      <div className="flex gap-4 text-xs text-muted">
        {labels.map((l, i) => (
          <span key={i} className="flex items-center gap-1.5">
            <span className={`h-2.5 w-2.5 rounded-sm ${segment[i]!.split(" ")[0]}`} />
            {l}
          </span>
        ))}
      </div>
    </div>
  );
}
