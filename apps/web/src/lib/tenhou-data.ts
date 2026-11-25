import "server-only";
import { cacheLife } from "next/cache";
import { rest } from "./supabase-rest";

export interface PlayerSummary {
  id: number;
  display_name: string;
  first_seen_at: string;
  last_seen_at: string;
}

export interface GameRow {
  placement: number;
  result_pt: number;
  rank_pt_delta: number | null;
  rank_before: number | null;
  final_score: number | null;
  games: {
    started_at: string;
    duration_min: number | null;
    has_full_log: boolean;
    rulesets: { players: 3 | 4; length: "east" | "south"; lobby: string; speed: string };
  };
}

export interface ModeStats {
  players: 3 | 4;
  games: number;
  placements: number[]; // counts per placement
  totalPt: number;
}

export interface TenhouPlayer {
  player: PlayerSummary;
  recent: GameRow[];
  modes: ModeStats[];
  sampled: number; // games used for stats (capped)
}

const STATS_LIMIT = 1000;

export async function searchTenhouPlayers(q: string): Promise<PlayerSummary[]> {
  "use cache";
  cacheLife("minutes");
  const term = q.trim();
  if (!term) return [];
  // PostgREST ilike: escape the pattern characters the user typed.
  const pattern = encodeURIComponent(`${term.replace(/[%_*]/g, "")}*`);
  const { data } = await rest<PlayerSummary[]>(
    `players?platform=eq.tenhou&platform_key=ilike.${pattern}&select=id,display_name,first_seen_at,last_seen_at&order=last_seen_at.desc&limit=20`,
  );
  return data;
}

export async function getTenhouPlayer(name: string): Promise<TenhouPlayer | null> {
  "use cache";
  cacheLife("minutes");
  const { data: players } = await rest<PlayerSummary[]>(
    `players?platform=eq.tenhou&platform_key=eq.${encodeURIComponent(name)}&select=id,display_name,first_seen_at,last_seen_at`,
  );
  const player = players[0];
  if (!player) return null;
  const { data: rows } = await rest<GameRow[]>(
    `game_participants?player_id=eq.${player.id}` +
      `&select=placement,result_pt,rank_pt_delta,rank_before,final_score,games!inner(started_at,duration_min,has_full_log,rulesets!inner(players,length,lobby,speed))` +
      `&order=games(started_at).desc&limit=${STATS_LIMIT}`,
  );
  const modes = new Map<number, ModeStats>();
  for (const r of rows) {
    const p = r.games.rulesets.players;
    const m = modes.get(p) ?? { players: p, games: 0, placements: Array(p).fill(0), totalPt: 0 };
    m.games++;
    m.placements[r.placement - 1]!++;
    m.totalPt += Number(r.result_pt);
    modes.set(p, m);
  }
  return {
    player,
    recent: rows.slice(0, 50),
    modes: [...modes.values()].sort((a, b) => b.games - a.games),
    sampled: rows.length,
  };
}
