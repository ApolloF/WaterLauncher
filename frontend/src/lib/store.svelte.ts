import { api } from "./api";
import type { AppInfo, Game, MetaState, ScanState, Session, Settings } from "./types";
import { lastPlayed, played, title } from "./types";

export type FilterKind = "all" | "installed" | "notinstalled" | "favorites" | "recent" | "found" | "hidden";
export type Filter = { kind: FilterKind } | { kind: "source"; source: string };
export type Sort = "title" | "recent" | "added" | "playtime";

export interface Toast {
  id: number;
  text: string;
  tone: "info" | "error";
}

// Source groups shown in the sidebar, in this order.
export const SOURCE_GROUPS: { id: string; label: string; match: (g: Game) => boolean }[] = [
  { id: "steam", label: "Steam", match: (g) => g.source === "steam" },
  { id: "epic", label: "Epic", match: (g) => g.source === "epic" },
  { id: "gog", label: "GOG", match: (g) => g.source === "gog" },
  { id: "ea", label: "EA", match: (g) => g.source === "ea" },
  { id: "ubisoft", label: "Ubisoft", match: (g) => g.source === "ubisoft" },
  { id: "battlenet", label: "Battle.net", match: (g) => g.source === "battlenet" },
  { id: "xbox", label: "Xbox", match: (g) => g.source === "xbox" },
  { id: "unofficial", label: "Unofficial", match: (g) => g.unofficial },
  { id: "standalone", label: "Standalone", match: (g) => !g.unofficial && (g.source === "installer" || g.source === "shortcut") },
  { id: "folder", label: "Folders", match: (g) => !g.unofficial && g.source === "folder" },
];

const WEEK = 7 * 86400;

/** A game found in the last week and not played yet. */
export const isFresh = (g: Game) => !g.initial && !lastPlayed(g) && !played(g) && Date.now() / 1000 - g.addedAt < WEEK;

const norm = (s: string) => s.toLowerCase().normalize("NFKD").replace(/[^\p{L}\p{N}]+/gu, "");

class LibraryStore {
  games = $state<Game[]>([]);
  settings = $state<Settings | null>(null);
  info = $state<AppInfo | null>(null);
  scan = $state<ScanState>({ running: true, lastScan: 0, tookMs: 0, games: 0, added: 0, known: 0 });
  meta = $state<MetaState>({ running: false, done: 0, total: 0 });
  loaded = $state(false);
  /** The game being launched or played (or the last one). */
  session = $state<Session | null>(null);

  filter = $state<Filter>({ kind: "all" });
  sort = $state<Sort>("title");
  query = $state("");
  selectedId = $state<number | null>(null);
  toasts = $state<Toast[]>([]);

  /** Games the library views show at all: installed ones (or every one when
   * the setting says so), without hidden ones. */
  base = $derived.by(() => {
    const showAll = this.settings?.showNotInstalled ?? false;
    return this.games.filter((g) => !g.hidden && (g.installed || showAll));
  });

  visible = $derived.by(() => {
    const f = this.filter;
    let list: Game[];
    if (f.kind === "hidden") list = this.games.filter((g) => g.hidden);
    else if (f.kind === "source") {
      const group = SOURCE_GROUPS.find((s) => s.id === f.source);
      list = this.base.filter((g) => (group ? group.match(g) : true));
    } else {
      list = this.base.filter((g) => {
        switch (f.kind) {
          case "installed":
            return g.installed;
          case "notinstalled":
            return !g.installed;
          case "favorites":
            return !!g.favorite;
          case "recent":
            return lastPlayed(g) > 0;
          case "found":
            return g.needsReview || isFresh(g);
          default:
            return true;
        }
      });
    }
    const q = norm(this.query);
    if (q) list = list.filter((g) => norm(title(g)).includes(q) || norm(g.sourceLabel).includes(q));
    const sort = f.kind === "recent" ? "recent" : this.sort;
    const byTitle = (a: Game, b: Game) => a.sortTitle.localeCompare(b.sortTitle);
    return [...list].sort((a, b) => {
      switch (sort) {
        case "recent":
          return lastPlayed(b) - lastPlayed(a) || byTitle(a, b);
        case "added":
          return b.addedAt - a.addedAt || byTitle(a, b);
        case "playtime":
          return played(b) - played(a) || byTitle(a, b);
        default:
          return byTitle(a, b);
      }
    });
  });

  selected = $derived.by(() => {
    const id = this.selectedId;
    return this.visible.find((g) => g.id === id) ?? this.visible[0] ?? null;
  });

  counts = $derived.by(() => {
    const b = this.base;
    return {
      all: b.length,
      installed: b.filter((g) => g.installed).length,
      notinstalled: b.filter((g) => !g.installed).length,
      favorites: b.filter((g) => g.favorite).length,
      recent: b.filter((g) => lastPlayed(g) > 0).length,
      found: b.filter((g) => g.needsReview || isFresh(g)).length,
      hidden: this.games.filter((g) => g.hidden).length,
    };
  });

  sources = $derived.by(() =>
    SOURCE_GROUPS.map((s) => ({ ...s, count: this.base.filter(s.match).length })).filter((s) => s.count > 0),
  );

  async init() {
    api.onLibraryChanged(() => void this.refresh());
    api.onScanState((s) => (this.scan = s));
    api.onMetaState((s) => (this.meta = s));
    api.launch.onSession((s) => (this.session = s));
    const [games, settings, scan, info, meta, session] = await Promise.all([
      api.games(),
      api.settings(),
      api.scanState(),
      api.info(),
      api.metaState(),
      api.launch.session(),
    ]);
    this.meta = meta;
    this.session = session;
    this.games = games;
    this.settings = settings;
    this.scan = scan;
    this.info = info;
    this.loaded = true;
  }

  async refresh() {
    try {
      this.games = await api.games();
    } catch (e) {
      this.toast(String(e), "error");
    }
  }

  /** Starts a game; the session events tell how it goes. */
  play(g: Game) {
    return this.run(() => api.launch.play(g.id));
  }

  /** Puts an updated game in place without waiting for a full refresh. */
  replace(g: Game) {
    const i = this.games.findIndex((x) => x.id === g.id);
    if (i >= 0) this.games[i] = g;
  }

  async saveSettings(next: Settings) {
    try {
      this.settings = await api.saveSettings($state.snapshot(next) as Settings);
    } catch (e) {
      this.toast(errText(e), "error");
    }
  }

  /** Runs an action, showing its error as a toast. */
  async run<T>(fn: () => Promise<T>): Promise<T | undefined> {
    try {
      const r = await fn();
      if (r && typeof r === "object" && "id" in r && "key" in r) this.replace(r as unknown as Game);
      return r;
    } catch (e) {
      this.toast(errText(e), "error");
      return undefined;
    }
  }

  private nextToast = 1;
  toast(text: string, tone: Toast["tone"] = "info") {
    const t = { id: this.nextToast++, text, tone };
    this.toasts = [...this.toasts, t];
    setTimeout(() => (this.toasts = this.toasts.filter((x) => x.id !== t.id)), tone === "error" ? 7000 : 3500);
  }
}

/** Error text without the Go/wails wrapping. */
export function errText(e: unknown): string {
  const s = e instanceof Error ? e.message : typeof e === "string" ? e : JSON.stringify(e);
  try {
    const j = JSON.parse(s);
    if (j && typeof j.message === "string") return j.message;
  } catch {
    /* not JSON */
  }
  return s;
}

export const lib = new LibraryStore();
