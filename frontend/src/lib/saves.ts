// A game's saves in words, from what Syncer reports.
import { ago } from "./format";
import type { Saves } from "./types";

export type SavesTone = "ok" | "warn" | "muted";

export interface SavesSummary {
  /** Short line: "Synced · backed up 2 h ago". */
  text: string;
  /** Longer explanation for details views. */
  detail: string;
  tone: SavesTone;
  /** What the button next to it does, if anything. */
  action?: "get" | "open";
}

const unix = (iso: string) => {
  const t = Date.parse(iso);
  return Number.isNaN(t) || t <= 0 ? 0 : Math.floor(t / 1000);
};

function backedUpText(iso: string): string {
  const t = unix(iso);
  return t ? `backed up ${ago(t).toLowerCase()}` : "not backed up yet";
}

export function savesSummary(s: Saves | null): SavesSummary | null {
  if (!s) return null;
  if (!s.installed)
    return {
      text: "Syncer isn't installed",
      detail: "Syncer keeps saves in sync between your PCs and backs them up. WaterLauncher then syncs them before you play.",
      tone: "muted",
      action: "get",
    };
  if (s.outdated)
    return {
      text: "Syncer needs an update",
      detail: "WaterLauncher works with Syncer 0.11 or newer: it then syncs a game's saves before you play and backs them up after.",
      tone: "warn",
      action: "get",
    };
  if (!s.available) return { text: "Syncer didn't answer", detail: s.error || "Try again in a moment.", tone: "warn", action: "open" };
  if (!s.known) return { text: "Not in Syncer", detail: "Syncer has no saves for this game (yet). It picks up new games on its own.", tone: "muted", action: "open" };

  const conflicts = s.folders.reduce((n, f) => n + f.conflicts, 0);
  if (conflicts)
    return {
      text: conflicts === 1 ? "2 versions of a save" : `${conflicts} saves with 2 versions`,
      detail: "Two PCs changed the same save. Pick the version to keep in Syncer before you play.",
      tone: "warn",
      action: "open",
    };
  const newer = s.folders.find((f) => f.newerOn);
  if (newer)
    return {
      text: `Newer save on ${newer.newerOn}`,
      detail: `${newer.newerOn} backed up a newer save that hasn't reached this PC yet. Turn that PC on, or wait for it when you play.`,
      tone: "warn",
      action: "open",
    };

  const synced = s.folders.filter((f) => f.sync);
  const latest = s.folders.map((f) => f.backedUp).sort().at(-1) ?? "";
  if (!synced.length) {
    const b = backedUpText(latest);
    return { text: `${b[0].toUpperCase()}${b.slice(1)} · not synced`, detail: "Syncer backs these saves up but doesn't sync them to other PCs.", tone: "ok" };
  }
  const state = synced.some((f) => f.state === "paused")
    ? "Syncing paused"
    : synced.some((f) => f.state !== "idle" || f.needBytes > 0)
      ? "Syncing…"
      : "Synced";
  return {
    text: `${state} · ${backedUpText(latest)}`,
    detail: "Syncer syncs these saves with your other PCs and backs them up. They're brought up to date before you play and backed up after.",
    tone: state === "Synced" ? "ok" : "warn",
  };
}
