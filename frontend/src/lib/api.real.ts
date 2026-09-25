import { Events, Window } from "@wailsio/runtime";
import { LibraryService, SettingsService } from "../../bindings/github.com/ApolloF/WaterLauncher/internal/app";
import type { Api } from "./api";
import type { AppInfo, Game, MetaState, ScanState, Settings, StoreHit } from "./types";

// The generated bindings return the Go structs; their JSON matches ./types.
const g = (p: Promise<unknown>) => p as Promise<Game>;

export const realApi: Api = {
  games: () => LibraryService.Games() as Promise<unknown> as Promise<Game[]>,
  scanState: () => LibraryService.ScanState() as Promise<unknown> as Promise<ScanState>,
  rescan: () => LibraryService.Rescan(),
  setFavorite: (id, on) => g(LibraryService.SetFavorite(id, on)),
  setHidden: (id, on) => g(LibraryService.SetHidden(id, on)),
  setPadMode: (id, mode) => g(LibraryService.SetPadMode(id, mode)),
  rename: (id, title) => g(LibraryService.Rename(id, title)),
  confirmMatch: (id) => g(LibraryService.ConfirmMatch(id)),
  chooseExe: (id) => g(LibraryService.ChooseExe(id)),
  play: (id) => LibraryService.Play(id),
  openFolder: (id) => LibraryService.OpenFolder(id),
  metaState: () => LibraryService.MetaState() as Promise<unknown> as Promise<MetaState>,
  refreshMetadata: (id) => LibraryService.RefreshMetadata(id),
  searchSteam: (q) => LibraryService.SearchSteam(q).then((h) => (h ?? []) as unknown as StoreHit[]),
  setMatch: (id, appId, name) => g(LibraryService.SetMatch(id, appId, name)),

  settings: () => SettingsService.Get() as Promise<unknown> as Promise<Settings>,
  saveSettings: (s) => SettingsService.Save(s as never) as Promise<unknown> as Promise<Settings>,
  addFolder: () => SettingsService.AddFolder() as Promise<unknown> as Promise<Settings>,
  removeFolder: (p) => SettingsService.RemoveFolder(p) as Promise<unknown> as Promise<Settings>,
  autoFolders: () => SettingsService.AutoFolders().then((f) => f ?? []),
  info: () => SettingsService.Info() as Promise<unknown> as Promise<AppInfo>,
  openLog: () => SettingsService.OpenLog(),
  hasSteamGridDBKey: () => SettingsService.HasSteamGridDBKey(),
  setSteamGridDBKey: (k) => SettingsService.SetSteamGridDBKey(k),

  onLibraryChanged: (cb) => Events.On("library:changed", () => cb()),
  onScanState: (cb) => Events.On("scan:state", (e) => cb(e.data as unknown as ScanState)),
  onMetaState: (cb) => Events.On("meta:state", (e) => cb(e.data as unknown as MetaState)),

  window: {
    minimise: () => void Window.Minimise(),
    toggleMaximise: () => void Window.ToggleMaximise(),
    close: () => void Window.Close(),
  },
};
