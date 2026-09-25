// Made-up library for `npm run dev:mock`: the games from the design canvas,
// covering every way a game can be found.
import type { Api } from "./api";
import type { AppInfo, Game, ScanState, Settings } from "./types";

const now = Math.floor(Date.now() / 1000);
const day = 86400;

let nextId = 1;
function game(p: Partial<Game> & { title: string }): Game {
  const id = nextId++;
  return {
    id,
    key: `mock:${id}`,
    sortTitle: p.title.toLowerCase().replace(/^the /, ""),
    source: "steam",
    sourceLabel: "Steam",
    unofficial: false,
    installed: true,
    dir: `D:\\Games\\${p.title}`,
    how: "Steam library",
    matchHow: "Steam library",
    confidence: 100,
    needsReview: false,
    addedAt: now - 90 * day,
    seenAt: now,
    ...p,
  };
}

let games: Game[] = [
  game({ title: "Ember Crown", source: "installer", sourceLabel: "Unofficial · Goldberg", unofficial: true, emulator: "Goldberg", steamAppId: 1245620, how: "Game folder in D:\\Games (Steam emulator)", matchHow: "Steam AppID read from steam_settings", confidence: 95, playtime: 18 * 3600, lastPlayed: now - 3600, favorite: true, exe: "D:\\Games\\Ember Crown\\EmberCrown.exe", sizeBytes: 54e9 }),
  game({ title: "Hollow Tide", steamAppId: 413150, launchUri: "steam://rungameid/413150", playtime: 42 * 3600, lastPlayed: now - day, favorite: true, dir: "C:\\Program Files (x86)\\Steam\\steamapps\\common\\Hollow Tide", sizeBytes: 64e9 }),
  game({ title: "Neon Meridian", source: "installer", sourceLabel: "Repack · DODI", unofficial: true, repacker: "DODI", how: "Installed by a DODI repack", matchHow: "Matched by title", confidence: 85, playtime: 7 * 3600, lastPlayed: now - 3 * day, padMode: "steam", sizeBytes: 21e9 }),
  game({ title: "Starfall Protocol", source: "epic", sourceLabel: "Epic", how: "Epic Games library", matchHow: "Epic Games library", playtime: 64 * 3600, lastPlayed: now - 8 * day, sizeBytes: 38e9 }),
  game({ title: "Grimwald", source: "installer", sourceLabel: "Unofficial · EMPRESS", unofficial: true, emulator: "EMPRESS", steamAppId: 1149460, how: "Installer entry in Windows", matchHow: "Steam AppID read from steam_emu.ini", confidence: 95, playtime: 88 * 3600, lastPlayed: now - 15 * day, favorite: true, sizeBytes: 47e9 }),
  game({ title: "Quiet Harbor", source: "gog", sourceLabel: "GOG", gogId: "1207658924", how: "GOG Galaxy library", matchHow: "GOG Galaxy library", playtime: 12 * 3600, lastPlayed: now - 34 * day, sizeBytes: 9e9 }),
  game({ title: "Frostline", source: "xbox", sourceLabel: "Xbox", how: "Xbox app library", matchHow: "Xbox app library", playtime: 9 * 3600, lastPlayed: now - 40 * day, sizeBytes: 88e9 }),
  game({ title: "Iron Veil", source: "installer", sourceLabel: "Unofficial · RUNE", unofficial: true, emulator: "RUNE", repacker: "FitGirl", steamAppId: 1086940, how: "Installed by a FitGirl repack", matchHow: "Steam AppID read from steam_emu.ini", confidence: 95, addedAt: now - 3600, sizeBytes: 71e9 }),
  game({ title: "Sable Run", source: "folder", sourceLabel: "Folder", how: "Game folder in D:\\Games (Unity)", matchHow: "Not matched to a known game", confidence: 40, needsReview: true, addedAt: now - 2 * 3600, padMode: "steam", sizeBytes: 3e9 }),
  game({ title: "Lumen Drift", steamAppId: 620, launchUri: "steam://rungameid/620", addedAt: now - day, sizeBytes: 6e9 }),
  game({ title: "Kestrel", source: "steam", installed: false, playtime: 6 * 3600, lastPlayed: now - 400 * day }),
  game({ title: "Copper Fields", source: "gog", sourceLabel: "GOG", installed: false, playtime: 21 * 3600, lastPlayed: now - 700 * day }),
];

let settings: Settings = {
  folders: ["D:\\Games"],
  autoFolders: true,
  detectUnofficial: true,
  reviewUncertain: true,
  showNotInstalled: false,
  theme: "system",
  bigPictureLayout: "deck",
  openBigPictureOnController: true,
  startInBigPicture: false,
  sounds: false,
  haptics: true,
  lightbar: true,
  psButton: true,
  glyphs: "auto",
};

let state: ScanState = { running: false, lastScan: now - 120, tookMs: 940, games: games.length, added: 0, known: 52107 };
const libListeners = new Set<() => void>();
const scanListeners = new Set<(s: ScanState) => void>();

const clone = <T>(v: T): T => JSON.parse(JSON.stringify(v));
const wait = (ms = 60) => new Promise((r) => setTimeout(r, ms));

function update(id: number, fn: (g: Game) => void): Promise<Game> {
  const g = games.find((x) => x.id === id);
  if (!g) return Promise.reject(new Error("game not found"));
  fn(g);
  libListeners.forEach((cb) => cb());
  return Promise.resolve(clone(g));
}

export const mockApi: Api = {
  async games() {
    await wait();
    return clone(games);
  },
  async scanState() {
    return clone(state);
  },
  async rescan() {
    state = { ...state, running: true };
    scanListeners.forEach((cb) => cb(clone(state)));
    await wait(1200);
    state = { ...state, running: false, lastScan: Math.floor(Date.now() / 1000) };
    scanListeners.forEach((cb) => cb(clone(state)));
    libListeners.forEach((cb) => cb());
  },
  setFavorite: (id, on) => update(id, (g) => (g.favorite = on)),
  setHidden: (id, on) => update(id, (g) => (g.hidden = on)),
  setPadMode: (id, mode) => update(id, (g) => (g.padMode = mode === "auto" ? "" : mode)),
  rename: (id, t) => update(id, (g) => (g.customTitle = t.trim())),
  confirmMatch: (id) => update(id, (g) => ((g.confirmed = true), (g.needsReview = false))),
  chooseExe: (id) => update(id, (g) => ((g.exe = g.dir + "\\Game.exe"), (g.userExe = true))),
  async play(id) {
    await update(id, (g) => (g.lastPlayed = Math.floor(Date.now() / 1000)));
  },
  async openFolder() {},

  async settings() {
    return clone(settings);
  },
  async saveSettings(s) {
    settings = clone(s);
    return clone(settings);
  },
  async addFolder() {
    settings = { ...settings, folders: [...settings.folders, "E:\\More Games"] };
    return clone(settings);
  },
  async removeFolder(p) {
    settings = { ...settings, folders: settings.folders.filter((f) => f !== p) };
    return clone(settings);
  },
  async autoFolders() {
    return ["C:\\Program Files (x86)\\DODI-Repacks", "D:\\Games"];
  },
  async info(): Promise<AppInfo> {
    return { version: "mock", dataDir: "C:\\Users\\you\\AppData\\Roaming\\WaterLauncher", logFile: "waterlauncher.log" };
  },
  async openLog() {},

  onLibraryChanged(cb) {
    libListeners.add(cb);
    return () => libListeners.delete(cb);
  },
  onScanState(cb) {
    scanListeners.add(cb);
    return () => scanListeners.delete(cb);
  },
  window: { minimise() {}, toggleMaximise() {}, close() {} },
};
