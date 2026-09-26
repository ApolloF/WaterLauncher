// Game lists and helpers shared by the big picture layouts.
import { artFor } from "./art";
import { isFresh } from "./store.svelte";
import { lastPlayed, played, title, type Game } from "./types";
import { playtime, ago } from "./format";
import { pad } from "./input.svelte";
import { padExplain } from "./route";

export const byRecent = (a: Game, b: Game) => lastPlayed(b) - lastPlayed(a) || a.sortTitle.localeCompare(b.sortTitle);
export const byTitle = (a: Game, b: Game) => a.sortTitle.localeCompare(b.sortTitle);

/** Installed games, most recently played first (then A–Z). */
export function recentFirst(games: Game[]): Game[] {
  return games.filter((g) => g.installed).sort(byRecent);
}

/** Recently played games; falls back to the whole library when nothing was played. */
export function continuePlaying(games: Game[], n = 10): Game[] {
  const played = games.filter((g) => g.installed && lastPlayed(g) > 0).sort(byRecent);
  return (played.length ? played : games.filter((g) => g.installed).sort(byTitle)).slice(0, n);
}

/** New finds and games that need a check, newest first. */
export function newFinds(games: Game[], n = 10): Game[] {
  return games
    .filter((g) => g.installed && (isFresh(g) || g.needsReview))
    .sort((a, b) => b.addedAt - a.addedAt)
    .slice(0, n);
}

export const accentOf = (g: Game | null | undefined) => (g ? (g.meta?.accent ?? artFor(g.key).accent) : undefined);

/** "18 h · Played yesterday" style line. */
export function metaLine(g: Game): string {
  const p = played(g);
  const parts = [p ? playtime(p) : "Not played yet"];
  const lp = lastPlayed(g);
  parts.push(lp ? `Played ${ago(lp).toLowerCase()}` : `Added ${ago(g.addedAt).toLowerCase()}`);
  return parts.join(" · ");
}

export { title };

/** What the controller mode will do, in words, for the big picture cards. */
export function padSummary(g: Game): { short: string; long: string } {
  return padExplain(g, pad);
}
