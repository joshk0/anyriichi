# Research: database hosting cost comparison (Q-071)

> Prices checked 2026-11-25 on official pricing pages. **[V]** verified on page, **Calc** computed from verified unit prices, **[U]** unverified.

## Workload assumptions
- Ingest 30k Tenhou games/day → 120k participant rows/day (~11M games, ~45M rows/yr). Postgres ≈ 5 GB Y1, ≈ 15 GB Y3 incl. indexes; Go worker batches every 20 min.
- 50k user-submitted full logs/yr (per-hand rows) + ~1 GB/yr raw logs in object storage.
- 1k–10k player-page views/day; filtered aggregates over 100–20,000 games per player; name-prefix search.
- ~20 KB JSON/view → 0.6–6 GB/mo DB→Vercel egress (estimate).

## Cost table ($/mo)

| Option | Y1 (~5 GB) | Y3 (~15 GB) | Notes |
|---|---|---|---|
| Supabase Free | n/a | n/a | 500 MB DB cap [V]; pauses after 1 week idle |
| **Supabase Pro** | **$25** | **~$31–32** | $10 compute credit covers Micro (1 GB) [V]; 8 GB disk incl, then $0.125/GB [V]; Micro recommended ≤10 GB → Small $0.0206/hr ≈ +$5 [V]; 250 GB egress, 100 GB file storage incl [V] |
| Cloud SQL db-f1-micro (0.6 GB) | ~$10–11 | not advisable | $0.0105/hr [V]; SSD ≈ $0.17/GiB-mo [V]; internet egress $0.19/GiB [V]; no SLA, no free tier |
| Cloud SQL db-g1-small (1.7 GB) | ~$28–29 | ~$31–33 | $0.035/hr [V]; no auth/API layer |
| **Self-hosted Postgres on e2-micro** | **~$0–1** | **~$0–1** | Always Free VM + 30 GB pd-standard [V]; Premium egress 1 GB free then $0.12/GiB [V]; Standard tier 200 GiB/mo free [V] (applies to free VM — inferred, confirm); backups → GCS 5 GB free [V] |
| Firestore Standard | ~$8–40 | ~$15–60 | see below |
| Neon Free | n/a | n/a | 1 GB [V] |
| Neon Launch | ~$21 | ~$25–44 | $0.106/CU-hr [V], never scales to zero under this load; storage $0.35/GB-mo [V] |
| Cloudflare D1 | ~$5–40 | **breaks** | 10 GB per-DB hard cap [V]; per-row-read billing |
| Turso | $4.99–~60 | ~$25–30 | per-row-read billing [V] |
| PlanetScale Postgres PS-5 | ~$5+storage | — | 512 MiB RAM [V]; too small |
| Xata | ~$10.4 | ~$13.2 | Micro $0.012/hr, $0.28/GB [V] |
| BigQuery side store | $0 | ~$0.1 | 10 GiB + 1 TiB query free [V]; nightly batch only |

## Firestore detail
- Free/day: 1 GiB stored, 50k reads, 20k writes, 20k deletes; 10 GiB/mo egress [V]. us-central1: reads $0.03/100k, writes $0.09/100k, storage ≈ $0.15/GiB-mo [V]. Aggregations bill 1 read per 1,000 index entries [V].
- Writes: base 150k/day → ~$3.5/mo; + precomputed aggregates (~360k/day) → ~$13/mo (Calc).
- **Storage is the trap:** auto-indexes make a participant doc ~0.5–1 KB vs ~100 B in Postgres → Y1 25–55 GiB ($4–8), Y3 75–165 GiB ($11–25) (Calc).
- Reads: live aggregation 100–250 reads/view → $9–22/mo at 10k views/day; precomputed ~5 reads/view → within free tier.
- **Poor fit:** composite index per filter combination; no GROUP BY (placement distribution = 4 count queries); multi-field derived stats can't be computed in-DB; rollups can't answer arbitrary date ranges; prefix search only case-sensitive range queries.
- **Browser-direct security:** rules evaluate Firebase Auth `request.auth`; Supabase JWTs not accepted → server route minting Firebase custom tokens needed. Public stats could be `allow read: if true`; worker writes via Admin SDK.

## Gotchas
- Supabase: disk can grow but not shrink; each extra project pays full compute [V].
- e2-micro self-host: 1 GB RAM shared with worker (shared_buffers ~128–256 MB); pd-standard ~0.75 read IOPS/GB → ~20 IOPS at 30 GB [U]; OS uses ~10 GB of the 30 GB; you own patching, backups, TLS, upgrades, single point of failure.
- Neon scale-to-zero doesn't help with steady traffic + 20-min batches.
- D1 single-threaded per DB, Workers/HTTP only; scanned index rows count as rows read (~$35/mo overage at 10k views).
- Vercel functions default to iad1 (us-east-1) [U] → place DB in us-east-1.

## Browser-direct access summary
Supabase RLS works natively with Supabase Auth (public-read stats, service-role writes). Firestore needs Firebase Auth/custom tokens. Others: server-side only — which is the better pattern anyway (caching/ISR).

## Recommendation (agent + Claude)
**Supabase Pro (~$25 → ~$31/mo)**, us-east-1, worker writes over TLS (ingress, free); per-player monthly rollup table maintained by the worker; ISR caching (5–15 min); Micro → Small at ~10 GB.
**Budget alternative (~$0–1/mo, prototype):** Postgres on the e2-micro beside the worker (Standard network tier, pg_dump to GCS, server-side access only); migrate to Supabase Pro when traffic matters — both plain Postgres. Note: Supabase Auth still works with a self-hosted DB (auth is a separate free Supabase project; Free tier's 500 MB limit is irrelevant for auth-only, but Free projects pause after 1 week idle — auth traffic keeps it active).
**Avoid:** Cloud SQL (pricier, $0.19/GiB egress), D1 (10 GB cap), Firestore (storage bloat, no GROUP BY, auth friction), Neon (≈ Supabase Pro cost without auth/storage).

## Sources
supabase.com/pricing · supabase.com/docs/guides/platform/compute-and-disk · supabase.com/docs/guides/platform/billing-on-supabase · cloud.google.com/sql/pricing · cloud.google.com/vpc/network-pricing · docs.cloud.google.com/free/docs/free-cloud-features · cloud.google.com/firestore/pricing · firebase.google.com/docs/firestore/pricing · cloud.google.com/bigquery/pricing · neon.com/pricing · developers.cloudflare.com/d1/platform/pricing · developers.cloudflare.com/d1/platform/limits · turso.tech/pricing · planetscale.com/pricing · xata.io/pricing
