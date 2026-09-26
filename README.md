<p align="center"><img src="build/appicon.png" width="96" alt=""></p>

<h1 align="center">WaterLauncher</h1>

<p align="center">A Windows game launcher that finds every game on your PC on its own, store installs and unofficial copies alike, and plays great with a DualSense.</p>

---

**[Download WaterLauncher](https://github.com/ApolloF/WaterLauncher/releases/latest/download/WaterLauncher-setup.exe)** for Windows 10 and 11 (64-bit). It installs for your account only, without administrator rights, and keeps itself up to date. Release notes are on the [Releases](https://github.com/ApolloF/WaterLauncher/releases) page; the plan is in [docs/PLAN.md](docs/PLAN.md).

- Finds Steam, Epic, GOG, EA, Ubisoft, Battle.net and Xbox installs, plus emulated-Steam copies, repacks, GOG rips and plain game folders, and works out which game each one is.
- Desktop mode for mouse and keyboard, and a big picture mode for controllers with three layouts to choose from (Deck, Console, Orbit).
- Starts games and tracks playtime. While you play, the interface closes to free memory, and the PS button opens an overlay over the game.
- DualSense first: native button glyphs, haptics, lightbar, and the PS button to open the launcher. Games without DualSense support start through Steam Input automatically.
- Works with [Syncer](https://github.com/ApolloF/syncer) to keep saves in sync before and after you play, and supports add-ons such as [DLSS Updater](https://github.com/ApolloF/dlssupdater).

WaterLauncher only manages games that are already installed. It never downloads games.

## Install and update

- **Installer:** `WaterLauncher-setup.exe` installs to `%LOCALAPPDATA%\Programs\WaterLauncher` with Start menu and desktop shortcuts. Uninstall from *Settings → Apps*; it asks before deleting your library.
- **Without installing:** `WaterLauncher.exe` from the same release runs from any folder.
- **Updates:** checked on GitHub twice a day (*Settings → General*). New versions download in the background and install the next time WaterLauncher starts. Each one must be signed with a release key that never leaves the maintainer's PC ([docs/RELEASING.md](docs/RELEASING.md)).
- **Start with Windows:** *Settings → General*. WaterLauncher then waits in the tray.
- **Games started elsewhere:** start a library game from Steam or a shortcut and WaterLauncher still counts its playtime and lets go of the controller (*Settings → Big picture → While playing*).
- **Command line:** `--play <id>` starts a game without the interface (for shortcuts), `--tray` starts in the tray, `--quit` closes a running WaterLauncher.
- **Your data:** `%APPDATA%\WaterLauncher` (library, settings, encrypted keys, log) and `%LOCALAPPDATA%\WaterLauncher` (art, the game database, add-ons, updates).

Builds aren't code-signed yet, so SmartScreen may warn the first time: *More info → Run anyway*. See [SECURITY.md](SECURITY.md) for how WaterLauncher keeps you safe.

## Develop

Requirements: Go 1.27+, Node 24+, [Wails v3](https://v3.wails.io) (`go install github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0-beta.26`).

```bash
wails3 build                              # bin/WaterLauncher.exe
wails3 task installer VERSION=v1.0.0 MAKENSIS="C:/Program Files (x86)/NSIS/makensis.exe"   # bin/WaterLauncher-setup.exe
go test ./...                             # backend tests
cd frontend && npm run check              # type-check the interface
cd frontend && npm run dev:mock           # interface in a browser with made-up games
WL_REAL_SCAN=1 go test -run RealScan -v ./internal/scan   # scan this PC and print what was found
```

The backend lives in `internal/`: `scan` (sources, unofficial-copy detection, executable picking), `identify` (Ludusavi matching), `library` (the game library), `settings`, `meta` (metadata and art), `pad` (controllers through SDL3), `launch` (game sessions: hooks, process tracking, playtime), `syncer` (client for Syncer's launcher API), `addons` (add-on host: manifests, approvals, processes), `owned` (owned games from Steam, GOG and Epic accounts), `update` (updates from GitHub releases), `sqlite` (read-only SQLite reader for GOG Galaxy), `steaminput` (Steam shortcuts for the Steam Input route), `app` (services the interface calls, windows and tray), plus small helpers (`lnk`, `platform`, `logx`). The installer is `build/windows/nsis/project.nsi`; signing is described in [docs/SIGNING.md](docs/SIGNING.md). Steam and game-database parsing comes from [gamekit](https://github.com/ApolloF/gamekit). The Svelte 5 interface is in `frontend/src`: `desktop/` for desktop mode, `bigpicture/` for the controller layouts, `overlay/` for the in-game overlay.

## License

[MIT](LICENSE)
