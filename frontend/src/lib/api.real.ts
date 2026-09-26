import { Events, Window } from "@wailsio/runtime";
import { LaunchService, LibraryService, PadService, SavesService, SettingsService } from "../../bindings/github.com/ApolloF/WaterLauncher/internal/app";
import type { Api } from "./api";
import type { AppInfo, Game, MetaState, PadState, Saves, ScanState, Session, Settings, StoreHit } from "./types";

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

  saves: {
    get: (id, fresh = false) => SavesService.Saves(id, fresh) as Promise<unknown> as Promise<Saves>,
    openSyncer: () => SavesService.OpenSyncer(),
    getSyncer: () => SavesService.GetSyncer(),
  },

  launch: {
    play: (id) => LaunchService.Play(id),
    session: () => LaunchService.Session() as Promise<unknown> as Promise<Session>,
    skip: (id) => void LaunchService.Skip(id),
    answer: (q, o) => void LaunchService.Answer(q, o),
    cancel: () => void LaunchService.Cancel(),
    quitGame: () => LaunchService.QuitGame(),
    setUIMode: (m) => void LaunchService.SetUIMode(m),
    closeOverlay: () => void LaunchService.CloseOverlay(),
    openMain: () => void LaunchService.OpenMain(),
    onSession: (cb) => Events.On("launch:session", (e) => cb(e.data as unknown as Session)),
    onOverlayAction: (cb) =>
      Events.On("overlay:action", (e) => {
        const d = e.data as unknown as { action: string; repeat: boolean };
        cb(d.action, d.repeat);
      }),
    onUIMode: (cb) => Events.On("ui:mode", (e) => cb(e.data as unknown as "desktop" | "bigpicture")),
  },

  window: {
    minimise: () => void Window.Minimise(),
    toggleMaximise: () => void Window.ToggleMaximise(),
    close: () => void Window.Close(),
    fullscreen: (on) => void (on ? Window.Fullscreen() : Window.UnFullscreen()),
  },

  pad: {
    state: () => PadService.State() as Promise<unknown> as Promise<PadState>,
    rumble: (effect) => void PadService.Rumble(effect),
    setLight: (hex) => void PadService.SetLight(hex),
    onAction: (cb) =>
      Events.On("pad:action", (e) => {
        const d = e.data as unknown as { action: string; repeat: boolean };
        cb(d.action, d.repeat);
      }),
    onState: (cb) => Events.On("pad:state", (e) => cb(e.data as unknown as PadState)),
  },
};
