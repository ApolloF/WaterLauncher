// The frontend's one door to the Go side. In mock mode (`npm run dev:mock`)
// the same interface is served by made-up data, so the interface can be
// built and checked in a normal browser. Vite drops the unused one.
import type { AddonGame, AddonView, AppInfo, Game, MetaState, PadState, Saves, ScanState, Session, Settings, StoreHit } from "./types";
import { realApi } from "./api.real";
import { mockApi } from "./api.mock";

export interface Api {
  games(): Promise<Game[]>;
  scanState(): Promise<ScanState>;
  rescan(): Promise<void>;
  setFavorite(id: number, on: boolean): Promise<Game>;
  setHidden(id: number, on: boolean): Promise<Game>;
  setPadMode(id: number, mode: string): Promise<Game>;
  rename(id: number, title: string): Promise<Game>;
  confirmMatch(id: number): Promise<Game>;
  chooseExe(id: number): Promise<Game>;
  openFolder(id: number): Promise<void>;
  metaState(): Promise<MetaState>;
  refreshMetadata(id: number): Promise<void>;
  searchSteam(query: string): Promise<StoreHit[]>;
  setMatch(id: number, appId: number, name: string): Promise<Game>;

  settings(): Promise<Settings>;
  saveSettings(s: Settings): Promise<Settings>;
  addFolder(): Promise<Settings>;
  removeFolder(path: string): Promise<Settings>;
  autoFolders(): Promise<string[]>;
  info(): Promise<AppInfo>;
  openLog(): Promise<void>;
  hasSteamGridDBKey(): Promise<boolean>;
  setSteamGridDBKey(key: string): Promise<void>;

  onLibraryChanged(cb: () => void): () => void;
  onScanState(cb: (s: ScanState) => void): () => void;
  onMetaState(cb: (s: MetaState) => void): () => void;

  saves: {
    /** Syncer's view of a game's saves; fresh skips the short cache. */
    get(id: number, fresh?: boolean): Promise<Saves>;
    openSyncer(): Promise<void>;
    getSyncer(): Promise<void>;
  };

  addons: {
    list(): Promise<AddonView[]>;
    /** Turns an add-on on, approving the program with this SHA-256. */
    enable(id: string, sha256: string): Promise<AddonView>;
    disable(id: string): Promise<AddonView>;
    add(): Promise<AddonView | null>;
    remove(id: string): Promise<void>;
    openFolder(): Promise<void>;
    forGame(id: number): Promise<AddonGame[]>;
    runAction(id: number, addon: string, action: string): Promise<string>;
    onProgress(cb: (addon: string, text: string) => void): () => void;
  };

  launch: {
    play(id: number): Promise<void>;
    session(): Promise<Session>;
    skip(stepId: string): void;
    answer(questionId: number, option: string): void;
    cancel(): void;
    quitGame(): Promise<void>;
    setUIMode(mode: "desktop" | "bigpicture"): void;
    closeOverlay(): void;
    openMain(): void;
    onSession(cb: (s: Session) => void): () => void;
    /** Controller actions while the in-game overlay shows. */
    onOverlayAction(cb: (action: string, repeat: boolean) => void): () => void;
    /** The tray asks for a mode. */
    onUIMode(cb: (mode: "desktop" | "bigpicture") => void): () => void;
  };

  window: {
    minimise(): void;
    toggleMaximise(): void;
    close(): void;
    fullscreen(on: boolean): void;
  };

  pad: {
    state(): Promise<PadState>;
    rumble(effect: "tick" | "confirm" | "error"): void;
    setLight(hex: string): void;
    onAction(cb: (action: string, repeat: boolean) => void): () => void;
    onState(cb: (s: PadState) => void): () => void;
  };
}

export const api: Api = import.meta.env.VITE_MOCK === "1" ? mockApi : realApi;
