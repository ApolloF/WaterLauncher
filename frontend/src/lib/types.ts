// Shapes of what the Go side sends. They mirror the JSON of
// internal/library.Game, internal/settings.Settings and friends.

export interface Meta {
  description?: string;
  developers?: string[];
  publishers?: string[];
  genres?: string[];
  releaseDate?: string;
  releaseYear?: number;
  dualSense?: string; // "yes" | "no" | ""
  controller?: string;
  cover?: string;
  hero?: string;
  logo?: string;
  icon?: string;
  accent?: string;
  fetchedAt?: number;
  source?: string;
}

export interface Game {
  id: number;
  key: string;
  title: string;
  customTitle?: string;
  sortTitle: string;
  source: string;
  sourceLabel: string;
  unofficial: boolean;
  emulator?: string;
  repacker?: string;
  drmFree?: string;
  installed: boolean;
  padHint?: string; // "libScePad" | "SDL"
  dir: string;
  exe?: string;
  args?: string;
  workDir?: string;
  launchUri?: string;
  userExe?: boolean;
  sizeBytes?: number;
  steamAppId?: number;
  metaAppId?: number;
  gogId?: string;
  epicApp?: string;
  how: string;
  matchHow: string;
  confidence: number;
  needsReview: boolean;
  confirmed?: boolean;
  addedAt: number;
  initial?: boolean;
  seenAt: number;
  lastPlayed?: number;
  playtime?: number;
  storePlaytime?: number;
  storeLastPlayed?: number;
  favorite?: boolean;
  hidden?: boolean;
  padMode?: string; // "" | "native" | "steam"
  meta?: Meta;
}

export type Layout = "deck" | "console" | "orbit";

export interface Settings {
  folders: string[];
  autoFolders: boolean;
  detectUnofficial: boolean;
  reviewUncertain: boolean;
  showNotInstalled: boolean;
  theme: "system" | "dark" | "light";
  bigPictureLayout: Layout;
  openBigPictureOnController: boolean;
  startInBigPicture: boolean;
  sounds: boolean;
  haptics: boolean;
  lightbar: boolean;
  psButton: boolean;
  glyphs: "auto" | "playstation" | "xbox";
  closeWhilePlaying: boolean;
  padWhilePlaying: "listen" | "off";
  syncSavesBefore: boolean;
  backupSavesAfter: boolean;
}

/** One save folder Syncer looks after. */
export interface SaveFolder {
  id: string;
  label: string;
  path: string;
  sync: boolean;
  backup: boolean;
  state: string;
  needBytes: number;
  errors: number;
  conflicts: number;
  exists: boolean;
  modified: string;
  backedUp: string;
  newerOn: string;
  newerAt: string;
}

/** What Syncer knows about a game's saves. */
export interface Saves {
  installed: boolean;
  outdated?: boolean;
  available: boolean;
  error?: string;
  known: boolean;
  folders: SaveFolder[];
}

export interface ScanState {
  running: boolean;
  lastScan: number;
  tookMs: number;
  games: number;
  added: number;
  known: number;
  error?: string;
}

export interface MetaState {
  running: boolean;
  done: number;
  total: number;
}

export interface StoreHit {
  appId: number;
  name: string;
  image: string;
}

export type PadKind = "playstation" | "xbox" | "nintendo" | "other";

export interface PadState {
  connected: boolean;
  name: string;
  kind: PadKind;
  dualSense: boolean;
  battery: number; // -1 unknown
  wireless: boolean;
  error?: string;
}

export type Phase = "preparing" | "starting" | "running" | "finishing" | "ended" | "failed" | "cancelled" | "";

export interface StepState {
  id: string;
  label: string;
  status: "pending" | "running" | "done" | "skipped" | "failed";
  detail?: string;
}

export interface Question {
  id: number;
  text: string;
  options: { id: string; label: string }[];
}

/** A game being launched or played (or the last one). */
export interface Session {
  id: number;
  gameId: number;
  title: string;
  phase: Phase;
  route: "direct" | "store" | "steamInput" | "";
  before: StepState[];
  after: StepState[];
  question?: Question;
  startedAt?: number;
  seconds: number;
  error?: string;
  note?: string;
}

export const sessionActive = (s: Session | null | undefined) =>
  !!s && s.phase !== "" && s.phase !== "ended" && s.phase !== "failed" && s.phase !== "cancelled";

export interface AppInfo {
  version: string;
  dataDir: string;
  logFile: string;
}

export const title = (g: Game) => g.customTitle || g.title;

/** Playtime in seconds: WaterLauncher's own or the store's, whichever is larger. */
export const played = (g: Game) => Math.max(g.playtime ?? 0, g.storePlaytime ?? 0);

/** When the game was last played, by WaterLauncher or the store. */
export const lastPlayed = (g: Game) => Math.max(g.lastPlayed ?? 0, g.storeLastPlayed ?? 0);
