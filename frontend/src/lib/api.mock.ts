// Made-up library for `npm run dev:mock`: the games from the design canvas,
// covering every way a game can be found.
import type { Api } from "./api";
import type { AddonGame, AddonView, AppInfo, Game, MetaState, Saves, ScanState, Session, Settings } from "./types";
import { sessionActive } from "./types";

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
  game({ meta: { description: "A fallen knight climbs a burning mountain to take back a crown that was never theirs. Brutal, fair combat and a world that remembers every choice.", developers: ["Ashgrove"], publishers: ["Ashgrove"], genres: ["Action", "RPG"], releaseYear: 2026, dualSense: "yes", accent: "#e8894a" }, title: "Ember Crown", source: "installer", sourceLabel: "Unofficial · Goldberg", unofficial: true, emulator: "Goldberg", steamAppId: 1245620, how: "Game folder in D:\\Games (Steam emulator)", matchHow: "Steam AppID read from steam_settings", confidence: 95, playtime: 18 * 3600, lastPlayed: now - 3600, favorite: true, exe: "D:\\Games\\Ember Crown\\EmberCrown.exe", sizeBytes: 54e9 }),
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
  closeWhilePlaying: true,
  padWhilePlaying: "listen",
  syncSavesBefore: true,
  backupSavesAfter: true,
};

const hour = 3600;
function mockSaves(g: Game | undefined): Saves {
  const base = { installed: true, available: true, known: false, folders: [] as Saves["folders"] };
  if (!g) return base;
  const iso = (s: number) => new Date(s * 1000).toISOString();
  const folder = (p: Partial<Saves["folders"][number]>) => ({
    id: g.key, label: g.title, path: "C:\\Users\\you\\AppData\\Roaming\\" + g.title, sync: true, backup: true, state: "idle",
    needBytes: 0, errors: 0, conflicts: 0, exists: true, modified: iso(now - 5 * hour), backedUp: iso(now - 2 * hour), newerOn: "", newerAt: "", ...p,
  });
  switch (g.title) {
    case "Ember Crown":
      return { ...base, known: true, folders: [folder({})] };
    case "Hollow Tide":
      return { ...base, known: true, folders: [folder({ conflicts: 2 })] };
    case "Grimwald":
      return { ...base, known: true, folders: [folder({ newerOn: "DESKTOP-TV", newerAt: iso(now - hour) })] };
    case "Quiet Harbor":
      return { ...base, known: true, folders: [folder({ sync: false, state: "backup-only" })] };
    case "Frostline":
      return { installed: false, available: false, known: false, folders: [] };
  }
  return base;
}

let sgdb = false;
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

// ---- a pretend add-on ----

let addon: AddonView = {
  id: "dlssupdater", name: "DLSS Updater", version: "1.4.0", publisher: "ApolloF",
  description: "Keeps DLSS up to date and installs OptiScaler.", homepage: "https://github.com/ApolloF/dlssupdater",
  exe: "C:\\Users\\you\\AppData\\Local\\Programs\\DLSS Updater\\DLSSUpdater.exe", hooks: ["game.status", "game.actions", "game.beforeLaunch"],
  permissions: [{ id: "modifyGameFiles", label: "Changes files in game folders" }, { id: "network", label: "Downloads from the internet" }],
  signed: false, sha256: "9f2c4e1ab07d5c3e88f1a6b2d4c90e7f13a5b8c2d6e0f4a7b9c1d3e5f7a9b2c4", enabled: false, running: false, state: "off",
};

function addonFor(g: Game | undefined): AddonGame[] {
  if (!addon.enabled || !g) return [];
  const hasDlss = ["Ember Crown", "Hollow Tide", "Starfall Protocol"].includes(g.title);
  return [{
    addon: addon.id, name: addon.name,
    badges: hasDlss ? [{ text: "DLSS 310.2", tone: "ok" }, ...(g.title === "Hollow Tide" ? [{ text: "Reverted by a patch", tone: "warn" as const }] : [])] : [{ text: "No DLSS" }],
    lines: hasDlss ? [{ label: "Super Resolution", value: "310.2.1" }, { label: "OptiScaler", value: "Not installed" }] : [],
    actions: hasDlss
      ? [
          { id: "update", label: "Update DLSS", description: "Install the latest DLSS files" },
          { id: "opti", label: "Install OptiScaler" },
          { id: "restore", label: "Restore original DLSS", confirm: "Put back the DLSS files the game shipped with?" },
          { id: "open", label: "Open in DLSS Updater" },
        ]
      : [{ id: "open", label: "Open in DLSS Updater" }],
  }];
}
const progressListeners = new Set<(a: string, t: string) => void>();

// ---- a pretend game session ----

let session: Session = { id: 0, gameId: 0, title: "", phase: "", route: "", before: [], after: [], seconds: 0 };
const sessionListeners = new Set<(s: Session) => void>();
let skipStep = "";
let answerWith: ((o: string) => void) | null = null;

function setSession(p: Partial<Session>) {
  session = { ...session, ...p };
  sessionListeners.forEach((cb) => cb(clone(session)));
}

async function runMockSession(g: Game) {
  const steam = g.padMode === "steam" && !g.launchUri;
  setSession({
    id: session.id + 1, gameId: g.id, title: g.customTitle || g.title, phase: "preparing",
    route: "", before: steam ? [{ id: "steamInput", label: "Steam Input", status: "running" }] : [],
    after: [], seconds: 0, startedAt: 0, error: "", note: "", question: undefined,
  });
  const id = session.id;
  if (steam) {
    const answer = await new Promise<string>((resolve) => {
      answerWith = resolve;
      setSession({ question: { id: 1, text: `Steam needs to restart once to add ${session.title} for Steam Input.`, options: [
        { id: "restart", label: "Restart Steam" }, { id: "direct", label: "Start without Steam Input" }, { id: "cancel", label: "Cancel" },
      ] } });
    });
    answerWith = null;
    setSession({ question: undefined });
    if (answer === "cancel") return setSession({ phase: "cancelled", before: [{ id: "steamInput", label: "Steam Input", status: "skipped" }] });
    setSession({ before: [{ id: "steamInput", label: "Steam Input", status: "running", detail: answer === "restart" ? "Closing Steam…" : "" }] });
    await wait(skipStep ? 0 : 1400);
    setSession({ before: [{ id: "steamInput", label: "Steam Input", status: "done", detail: answer === "restart" ? "Added to Steam" : "Starting without Steam Input" }] });
  }
  const cur = () => session; // read fresh after each await
  if (cur().id !== id || cur().phase !== "preparing") return;
  setSession({ phase: "starting", route: steam ? "steamInput" : g.launchUri ? "store" : "direct" });
  await update(g.id, (x) => (x.lastPlayed = Math.floor(Date.now() / 1000)));
  await wait(1800);
  if (cur().id !== id || cur().phase !== "starting") return;
  setSession({ phase: "running", startedAt: Math.floor(Date.now() / 1000) });
  const t = setInterval(() => {
    if (session.id !== id || session.phase !== "running") return clearInterval(t);
    setSession({ seconds: session.seconds + 1 });
  }, 1000);
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
  async openFolder() {},
  async metaState(): Promise<MetaState> {
    return { running: false, done: 0, total: 0 };
  },
  async refreshMetadata() {},
  async searchSteam(q) {
    await wait(300);
    return [
      { appId: 757310, name: "Sable", image: "" },
      { appId: 717850, name: `${q} Deluxe Edition`, image: "" },
      { appId: 941900, name: `${q} Soundtrack`, image: "" },
    ];
  },
  setMatch: (id, appId, name) => update(id, (g) => ((g.steamAppId = appId), (g.title = name), (g.confirmed = true), (g.needsReview = false), (g.matchHow = "Chosen by you"), (g.confidence = 100))),

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
  async hasSteamGridDBKey() {
    return sgdb;
  },
  async setSteamGridDBKey(k) {
    sgdb = !!k;
  },

  onLibraryChanged(cb) {
    libListeners.add(cb);
    return () => libListeners.delete(cb);
  },
  onScanState(cb) {
    scanListeners.add(cb);
    return () => scanListeners.delete(cb);
  },
  onMetaState() {
    return () => {};
  },
  saves: {
    async get(id) {
      await wait(250);
      return mockSaves(games.find((g) => g.id === id));
    },
    async openSyncer() {},
    async getSyncer() {},
  },
  addons: {
    async list() {
      return [clone(addon)];
    },
    async enable(_id, sha) {
      if (sha !== addon.sha256) throw new Error("the add-on's program changed while you were looking; check it again");
      addon = { ...addon, enabled: true, state: "on" };
      return clone(addon);
    },
    async disable() {
      addon = { ...addon, enabled: false, state: "off" };
      return clone(addon);
    },
    async add() {
      return null;
    },
    async remove() {},
    async openFolder() {},
    async forGame(id) {
      await wait(300);
      return addonFor(games.find((g) => g.id === id));
    },
    async runAction(_id, a, action) {
      for (const t of ["Downloading DLSS 310.2…", "Replacing files…"]) {
        progressListeners.forEach((cb) => cb(a, t));
        await wait(700);
      }
      return action === "open" ? "" : "Done";
    },
    onProgress(cb) {
      progressListeners.add(cb);
      return () => progressListeners.delete(cb);
    },
  },
  launch: {
    async play(id) {
      const g = games.find((x) => x.id === id);
      if (!g) throw new Error("game not found");
      if (sessionActive(session)) throw new Error(`${session.title} is still running`);
      runMockSession(g);
    },
    async session() {
      return clone(session);
    },
    skip(id) {
      skipStep = id;
    },
    answer(_q, option) {
      answerWith?.(option);
    },
    cancel() {
      if (session.phase === "preparing" || session.phase === "starting") setSession({ phase: "cancelled", question: undefined });
    },
    async quitGame() {
      if (session.phase !== "running") throw new Error("no game is running");
      setSession({ phase: "finishing", after: [{ id: "savesAfter", label: "Back up saves", status: "running", detail: "Backing up…" }] });
      await wait(1500);
      setSession({ phase: "ended", after: [{ id: "savesAfter", label: "Back up saves", status: "done", detail: "Backed up" }] });
    },
    setUIMode() {},
    closeOverlay() {},
    openMain() {},
    onSession(cb) {
      sessionListeners.add(cb);
      return () => sessionListeners.delete(cb);
    },
    onOverlayAction() {
      return () => {};
    },
    onUIMode() {
      return () => {};
    },
  },
  window: { minimise() {}, toggleMaximise() {}, close() {}, fullscreen() {} },
  pad: {
    async state() {
      return { connected: true, name: "DualSense Wireless Controller", kind: "playstation", dualSense: true, battery: 82, wireless: true };
    },
    rumble() {},
    setLight() {},
    onAction() {
      return () => {};
    },
    onState() {
      return () => {};
    },
  },
};
