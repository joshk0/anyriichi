/**
 * Ranked-game rules: normalized ±pt (D-016) and rank-point progression for Tenhou and Mahjong Soul.
 * Data lives in ../rules/*.json (sources cited there); the Go port is worker/internal/rules and both
 * run the vectors in ../testdata/rules.json — change them together.
 */
import tenhou from "../rules/tenhou.json";
import majsoul from "../rules/majsoul.json";
import type { TenhouLevel } from "./tenhouRule";

export type Platform = "tenhou" | "majsoul";
export type Players = 3 | 4;
export type Length = "east" | "south";
export type RankEvent = "promotion" | "demotion" | null;
export type Lang = "en" | "ja";

interface Scoring {
  startPoints: number;
  returnPoints: number;
  uma: number[];
}

function scoring(platform: Platform, players: Players): Scoring {
  const s = (platform === "tenhou" ? tenhou : majsoul).scoring[String(players) as "3" | "4"];
  if (!s) throw new Error(`no scoring for ${platform} ${players}p`);
  return s;
}

function checkPlacement(placement: number, players: Players): void {
  if (!Number.isInteger(placement) || placement < 1 || placement > players) {
    throw new Error(`invalid placement ${placement} for ${players} players`);
  }
}

/**
 * Normalized result in ±pt for one player: (score − return)/1000 + uma[placement], plus oka to 1st.
 * Placement is 1-based and must be the platform's own (both platforms break ties by seat order).
 * Rounded to 0.1.
 */
export function resultPt(platform: Platform, players: Players, score: number, placement: number): number {
  checkPlacement(placement, players);
  const s = scoring(platform, players);
  const okaTenths = placement === 1 ? ((s.returnPoints - s.startPoints) * players) / 100 : 0;
  // Work in tenths of a point (scores are multiples of 100) to avoid float drift.
  const tenths = Math.round((score - s.returnPoints) / 100) + s.uma[placement - 1]! * 10 + okaTenths;
  return tenths / 10;
}

// ---------------------------------------------------------------- Tenhou

export interface TenhouRank {
  rank: number; // index into tenhou.json ranks (0 = 新人 … 19 = 十段, 20 = 天鳳位)
  pt: number;
}

export interface TenhouGame {
  players: Players;
  level: TenhouLevel;
  length: Length;
  placement: number;
}

export interface RankChange<S> {
  before: S;
  after: S;
  delta: number;
  event: RankEvent;
}

const TENHOU_RANKS = tenhou.ranks;
export const TENHOU_SUPPORTED_FROM = new Date(tenhou.supportedFrom);

function tenhouRankAt(rank: number) {
  const r = TENHOU_RANKS[rank];
  if (!r) throw new Error(`invalid tenhou rank ${rank}`);
  return r;
}

export function tenhouRankName(rank: number, lang: Lang = "en"): string {
  return tenhouRankAt(rank)[lang];
}

/** Rank-point change for one ranked Tenhou game. */
export function tenhouRankDelta(rank: number, game: TenhouGame): number {
  checkPlacement(game.placement, game.players);
  const r = tenhouRankAt(rank);
  if (game.placement === game.players) return r.lastPlace[game.length];
  const table = tenhou.placementPoints[String(game.players) as "3" | "4"][game.level][game.length];
  return table[game.placement - 1]!;
}

/** Apply one game: promotion at ≥ promotePt, demotion below 0 (dan only); kyu ranks floor at 0. */
export function applyTenhouGame(state: TenhouRank, game: TenhouGame): RankChange<TenhouRank> {
  const r = tenhouRankAt(state.rank);
  if (r.promotePt === null) return { before: state, after: state, delta: 0, event: null }; // 天鳳位
  const delta = tenhouRankDelta(state.rank, game);
  const pt = state.pt + delta;
  if (pt >= r.promotePt) {
    const next = state.rank + 1;
    return { before: state, after: { rank: next, pt: tenhouRankAt(next).startPt }, delta, event: "promotion" };
  }
  if (pt < 0) {
    if (r.canDemote) {
      const prev = state.rank - 1;
      return { before: state, after: { rank: prev, pt: tenhouRankAt(prev).startPt }, delta, event: "demotion" };
    }
    return { before: state, after: { rank: state.rank, pt: 0 }, delta, event: null };
  }
  return { before: state, after: { rank: state.rank, pt }, delta, event: null };
}

/**
 * Whether a Tenhou log falls under the modelled rules (played on or after tenhou.json supportedFrom).
 * Older logs are declined (D-046). The log ID starts with the JST start hour, YYYYMMDDHH.
 */
export function tenhouLogSupported(logId: string): boolean {
  const m = /^(\d{4})(\d{2})(\d{2})(\d{2})gm-/.exec(logId);
  if (!m) return false;
  const started = new Date(`${m[1]}-${m[2]}-${m[3]}T${m[4]}:00:00+09:00`);
  return started.getTime() >= TENHOU_SUPPORTED_FROM.getTime();
}

// ---------------------------------------------------------------- Mahjong Soul

/** Amae Koromo-style level id: players digit (1 = 4p, 2 = 3p) × 10000 + major × 100 + minor. */
export interface MajsoulRank {
  levelId: number;
  points: number;
}

export interface MajsoulGame {
  mode: number; // Mahjong Soul room mode id, see majsoul.json modes
  score: number; // final score
  placement: number;
}

type MajsoulMode = (typeof majsoul.modes)[keyof typeof majsoul.modes] & { kontenDelta?: number[] };

const KONTEN = majsoul.konten;
const OLD_KONTEN_MAJOR = KONTEN.major - 1;
const KONTEN_FALLBACK_LEVEL = 503; // Saint 3: Konten playing outside Throne rooms scores like this

interface Level {
  p: number; // 1 = 4p, 2 = 3p
  major: number;
  minor: number;
}

const level = (id: number): Level => {
  const real = id % 10000;
  return { p: Math.floor(id / 10000), major: Math.floor(real / 100), minor: real % 100 };
};
const levelId = (l: Level): number => l.p * 10000 + l.major * 100 + l.minor;
const isKonten = (l: Level): boolean => l.major >= OLD_KONTEN_MAJOR;

function majsoulMode(mode: number): MajsoulMode {
  const m = (majsoul.modes as Record<string, MajsoulMode>)[String(mode)];
  if (!m) throw new Error(`unsupported majsoul mode ${mode}`);
  return m;
}

function maxPoints(l: Level): number {
  if (isKonten(l)) return l.minor === KONTEN.topMinor ? 0 : KONTEN.maxPoints;
  const v = majsoul.levelMaxPoints[(l.major - 1) * 3 + l.minor - 1];
  if (v === undefined) throw new Error(`invalid majsoul level ${levelId(l)}`);
  return v;
}

function startPoints(l: Level): number {
  return l.major === 1 ? 0 : maxPoints(l) / 2;
}

/** Pre-2022 Celestial (major 6) maps onto Konten level 1. */
function versionAdjusted(l: Level): Level {
  return l.major === OLD_KONTEN_MAJOR ? { p: l.p, major: KONTEN.major, minor: 1 } : l;
}

function versionAdjustedPoints(l: Level, points: number): number {
  return l.major === OLD_KONTEN_MAJOR ? Math.ceil(points / 100) * 10 + 200 : points;
}

function nextLevel(l0: Level): Level {
  const l = versionAdjusted(l0);
  let major = l.major;
  let minor = l.minor + 1;
  if (minor > 3 && !isKonten(l)) {
    major++;
    minor = 1;
  }
  if (major === OLD_KONTEN_MAJOR) major = KONTEN.major;
  return { p: l.p, major, minor };
}

function previousLevel(l0: Level): Level {
  if (l0.major === 1 && l0.minor === 1) return l0;
  const l = versionAdjusted(l0);
  let major = l.major;
  let minor = l.minor - 1;
  if (minor < 1) {
    major--;
    minor = 3;
  }
  if (major === OLD_KONTEN_MAJOR) major = OLD_KONTEN_MAJOR - 1;
  return { p: l.p, major, minor };
}

export function majsoulLevelName(id: number, lang: Lang = "en"): string {
  const l = level(id);
  const major = majsoul.majors.find((m) => m.major === l.major);
  if (!major) throw new Error(`invalid majsoul level ${id}`);
  if (l.major === OLD_KONTEN_MAJOR) return major[lang];
  return `${major[lang]} ${l.minor}`;
}

/** Rank-point change for one ranked Mahjong Soul game (last place includes the level penalty). */
export function majsoulRankDelta(id: number, game: MajsoulGame): number {
  let l = level(id);
  const mode = majsoulMode(game.mode);
  const players = mode.players as Players;
  checkPlacement(game.placement, players);
  const i = game.placement - 1;
  if (isKonten(l)) {
    if (mode.kontenDelta) return mode.kontenDelta[i]!;
    l = level(l.p * 10000 + KONTEN_FALLBACK_LEVEL);
  }
  const s = scoring("majsoul", players);
  // ceil((score − base)/1000 + uma), in integer arithmetic.
  let delta = Math.ceil((game.score - s.startPoints + s.uma[i]! * 1000) / 1000) + mode.placementBonus[i]!;
  if (game.placement === players) {
    const penalties = majsoul.lastPlacePenalty[mode.penaltyTable as keyof typeof majsoul.lastPlacePenalty];
    delta -= penalties[(l.major - 1) * 3 + l.minor - 1]!;
  }
  return delta + 0; // normalise -0
}

/** Apply one game, mirroring Mahjong Soul / Amae Koromo level adjustment. */
export function applyMajsoulGame(state: MajsoulRank, game: MajsoulGame): RankChange<MajsoulRank> {
  const delta = majsoulRankDelta(state.levelId, game);
  const original = level(state.levelId);
  let points = versionAdjustedPoints(original, state.points + delta);
  let l = versionAdjusted(original);
  const max = maxPoints(l);
  let event: RankEvent = null;
  if (max && points >= max) {
    l = nextLevel(l);
    points = startPoints(l);
    event = "promotion";
  } else if (points < 0) {
    if (!max || l.major === 1 || (l.major === 2 && l.minor === 1)) {
      points = 0;
    } else {
      l = previousLevel(l);
      points = startPoints(l);
      event = "demotion";
    }
  }
  return { before: state, after: { levelId: levelId(l), points }, delta, event };
}
