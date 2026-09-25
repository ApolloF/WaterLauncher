// The frontend's one door to the Go side. In mock mode (`npm run dev:mock`)
// the same interface is served by made-up data, so the interface can be
// built and checked in a normal browser. Vite drops the unused one.
import type { AppInfo, Game, ScanState, Settings } from "./types";
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
  play(id: number): Promise<void>;
  openFolder(id: number): Promise<void>;

  settings(): Promise<Settings>;
  saveSettings(s: Settings): Promise<Settings>;
  addFolder(): Promise<Settings>;
  removeFolder(path: string): Promise<Settings>;
  autoFolders(): Promise<string[]>;
  info(): Promise<AppInfo>;
  openLog(): Promise<void>;

  onLibraryChanged(cb: () => void): () => void;
  onScanState(cb: (s: ScanState) => void): () => void;

  window: {
    minimise(): void;
    toggleMaximise(): void;
    close(): void;
  };
}

export const api: Api = import.meta.env.VITE_MOCK === "1" ? mockApi : realApi;
