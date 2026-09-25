<p align="center"><img src="build/appicon.png" width="96" alt=""></p>

<h1 align="center">WaterLauncher</h1>

<p align="center">A Windows game launcher that finds every game on your PC on its own, store installs and unofficial copies alike, and plays great with a DualSense.</p>

---

**Status:** early previews. Download the latest prerelease from [Releases](https://github.com/ApolloF/WaterLauncher/releases). The plan is in [docs/PLAN.md](docs/PLAN.md).

- Finds Steam, Epic, GOG, EA, Ubisoft, Battle.net and Xbox installs, plus emulated-Steam copies, repacks, GOG rips and plain game folders, and works out which game each one is.
- Desktop mode for mouse and keyboard, and a big picture mode for controllers with three layouts to choose from (Deck, Console, Orbit).
- DualSense first: native button glyphs, haptics, lightbar, and the PS button to open the launcher. Games without DualSense support start through Steam Input automatically.
- Works with [Syncer](https://github.com/ApolloF/syncer) to keep saves in sync before and after you play, and supports add-ons such as [DLSS Updater](https://github.com/ApolloF/dlssupdater).

WaterLauncher only manages games that are already installed. It never downloads games.

## Develop

Requirements: Go 1.27+, Node 24+, [Wails v3](https://v3.wails.io) (`go install github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0-beta.26`).

```bash
wails3 build                              # bin/WaterLauncher.exe
go test ./...                             # backend tests
cd frontend && npm run check              # type-check the interface
cd frontend && npm run dev:mock           # interface in a browser with made-up games
WL_REAL_SCAN=1 go test -run RealScan -v ./internal/scan   # scan this PC and print what was found
```

The backend lives in `internal/`: `scan` (sources, unofficial-copy detection, executable picking), `identify` (Ludusavi matching), `library` (the game library), `settings`, `meta` (metadata and art), `pad` (controllers through SDL3), `app` (services the interface calls), plus small helpers (`lnk`, `vdf`, `platform`, `logx`). The Svelte 5 interface is in `frontend/src`: `desktop/` for desktop mode, `bigpicture/` for the controller layouts.

## License

[MIT](LICENSE)
