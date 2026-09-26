# WaterLauncher plan

Status: **v0.1 to v0.4 released** as prereleases (2026-09-25 and 26). **v0.5** (Syncer) is built (2026-09-26): [gamekit](https://github.com/ApolloF/gamekit) v0.1.0 is published and used by WaterLauncher, and the Syncer side is in [ApolloF/syncer#7](https://github.com/ApolloF/syncer/pull/7). v0.5 is released once Syncer 0.11 is out. v0.6 (add-ons) is next.
Design reference: [WaterLauncher Design Directions](https://claude.ai/artifact/EUqrFQcmgAThrxbm8vAr6i) (A Console, B Orbit, C Deck, D Desktop).

## 1. Goals

- A Windows game launcher that finds games on its own: store installs and unofficial copies (emulated Steam, repacks, GOG rips, plain folders), and adds art and metadata the way Playnite does for Steam games.
- Two modes: **Desktop** (mouse, layout D) and **Big picture** (controller). Big picture has three layouts, picked in Settings: **Deck (default)**, Console, Orbit.
- First-class DualSense: native input, glyphs, haptics, lightbar, and the PS button to summon the launcher. Games without DualSense support start through Steam Input automatically. Games that support it start directly.
- Syncer integration: save status per game, sync before launch, back up after exit, conflicts.
- Add-ons through hooks, starting with DLSS Updater.
- Pretty, responsive, secure and light: the interface unloads while you play.

Not in v1: downloading games or cracks (WaterLauncher only manages what's installed), emulators and ROMs, achievements, macOS or Linux, importing from Playnite.

## 2. Decisions

| Area | Decision |
|---|---|
| Stack | Go + Wails **v3** (pinned to `v3.0.0-beta.26`, upgraded on purpose, never floating) + Svelte 5 + TypeScript + Vite |
| Why v3 | It can close windows and keep the app alive (the interface unloads while a game runs), has several windows (main and in-game overlay) and a built-in tray. v2 can only hide its one window. |
| Library storage | In memory, saved as one JSON file with atomic writes (changed from SQLite during v0.1: even thousands of games stay a few MB, it loads in milliseconds, and it saves a ~7 MB dependency; search and filters run in the interface) |
| Controller | SDL3 3.4.x (`SDL3.dll`, zlib license) through a small binding of our own over `golang.org/x/sys/windows`. No cgo. We need about 20 functions, all with integer arguments. |
| Shared code | `github.com/ApolloF/gamekit` (public, MIT): VDF (text and binary), Steam (folder, libraries, installed apps, account, Authenticode, tamper and emulator checks) and the Ludusavi manifest parser. Store detection (Epic, GOG, Xbox, …) stays in each app for now: Syncer only needs names, WaterLauncher needs launch details. |
| Syncer | Stays a separate app and gets a local API (named pipe, current user only). Needs Syncer 0.11+; see [syncer-api.md](syncer-api.md) |
| Add-ons | Separate executables using JSON-RPC 2.0 over stdio. Their UI is declarative, so it renders natively in every layout. |
| Store games not installed | Optional, **off by default** |
| Desktop theme | Follows the Windows light or dark setting (Mica). Big picture is always dark. |
| License | MIT |
| Releases | Every phase version is published as a GitHub **prerelease** |

## 3. Architecture

```
                    ┌───────────────────────── WaterLauncher.exe (Go core, always running) ─────────────────────────┐
 metadata hosts ◄───┤ library (JSON)  scan  identify  meta+images  launch/track  pad (SDL3)  syncer client  addon host   │
 (allowlist)        └──────┬───────────────────────┬──────────────────────┬───────────────────────┬───────────────────────┘
                           │ Wails bindings/events │                      │ \\.\pipe\syncer       │ stdio JSON-RPC
                  ┌────────▼────────┐   ┌──────────▼─────────┐   ┌────────▼────────┐     ┌────────▼────────┐
                  │ Desktop window  │   │ Big picture window │   │ Syncer.exe      │     │ add-on (e.g.    │
                  │ (WebView2)      │   │ + overlay window   │   │ (separate app)  │     │ DLSS Updater)   │
                  └─────────────────┘   └────────────────────┘   └─────────────────┘     └─────────────────┘
                   windows are closed while a game runs; the Go core keeps tracking and listening for the PS button
```

Repo layout:

```
main.go                      app setup, windows, tray, single instance
internal/
  library/                   schema, migrations, queries, models
  scan/                      orchestrator, watchers
    sources/                 steam, epic, gog, xbox, ea, ubisoft, battlenet, uninstall, shortcuts, folders
    unofficial/              emulator and crack signatures -> store IDs
    exe/                     picking the main exe (ported from DLSS Updater's GameInspector)
  identify/                  ID match, Ludusavi, PCGamingWiki, fuzzy titles, confidence
  meta/                      steamstore, gog, pcgw, steamgriddb, images (fetch, validate, resize, cache, accent colour)
  launch/                    sessions: hooks pipeline, process tracking, playtime
  steaminput/                non-Steam shortcuts in shortcuts.vdf for the Steam Input route
  pad/                       SDL3 binding, DualSense features, intents, background PS-button listener
  syncer/                    pipe client (Syncer's launcher API)
  addons/                    manifest, host, permissions, JSON-RPC
  platform/                  DPAPI, known folders, foreground and full-screen checks
  update/                    GitHub release check, SHA-256 verify
frontend/src/
  lib/                       api (plus a mock backend), focus engine, input intents, glyphs, sounds, stores
  shared/                    GameArt, Logo, LaunchSequence, QuickAccess, Toggle, OSK
  desktop/                   layout D
  bigpicture/deck|console|orbit/
docs/                        PLAN.md, addon-protocol.md, syncer-api.md
```

Data: `%APPDATA%\WaterLauncher` (`settings.json`, `library.json`, log), `%LOCALAPPDATA%\WaterLauncher\cache` (game database, images, WebView2 data).

## 4. Library and detection

| Source | Signal | Identity |
|---|---|---|
| Steam | `libraryfolders.vdf`, `appmanifest_*.acf` | AppID |
| Epic | `ProgramData\Epic\...\Manifests\*.item` | catalog namespace and item |
| GOG | registry `GOG.com\Games`, `goggame-*.info` | GOG ID |
| EA, Ubisoft, Battle.net, Xbox | same as Syncer's `installed.go` | store IDs |
| Emulated Steam | `steam_appid.txt`, `steam_settings\` (Goldberg/GSE), `steam_emu.ini` (CODEX), RUNE/EMPRESS/TENOKE/SmartSteamEmu configs, `OnlineFix.ini`, an unsigned `steam_api(64).dll` | AppID from the files |
| GOG rips | `goggame-*.info` outside a GOG install | GOG ID |
| Repacks and installers | Uninstall registry entries (Inno Setup and others): `InstallLocation`, `DisplayIcon` | by title |
| Shortcuts | Start menu and desktop `.lnk` targets | by title |
| Watched folders | every subfolder is a candidate, live updates through `ReadDirectoryChangesW` | by title |

- **Identify:** exact ID first, then the Ludusavi manifest (install folder name to title and AppID, already used by Syncer), then PCGamingWiki, then normalised fuzzy title matching. Each match gets a confidence score. Low-confidence matches wait in *Found on this PC* for a check (setting on by default).
- **Main exe:** prefer the launcher exe in the game's root folder and known engine patterns; skip uninstallers, redistributables and crash handlers. Signals include version info and icon.
- **Duplicates:** the same folder or the same store ID merges. A store copy and an unofficial copy can both exist and are labelled.
- **Owned but not installed (opt-in):** Steam through the user's Web API key plus the SteamID from `loginusers.vdf`; GOG through Galaxy's local database when present; Epic through a sign-in the user completes themselves (token stored with DPAPI). *Install* opens the store client (`steam://install/…` and equivalents).
- **Speed:** the first scan runs in parallel, and later scans are incremental, driven by watchers.

## 5. Metadata and art

- **Steam (no key; checked 2026-09-25):** `store.steampowered.com/api/appdetails` gives the description, genres, developer, release date and categories (55 DualShock, 57/58 DualSense, 28 full controller support). `IStoreBrowseService/GetItems` with `include_assets` gives the library cover, hero, logo and header. This covers every emulated-Steam copy too.
- **GOG:** `api.gog.com/products/{id}`. **SteamGridDB:** optional key, for games without a Steam ID and to pick alternative covers. Games the manifest doesn't know are also looked up on the Steam store by exact title.
- **PCGamingWiki** was dropped during v0.2: its API now refuses Cargo queries. Controller support comes from Steam's store categories (DualShock 55, DualSense 57/58) and from the game's files (Sony's `libScePad.dll`, or SDL, which handles a DualSense itself).
- **Fallback:** a generated cover from the title plus the icon pulled from the exe.
- **Image pipeline:** HTTPS to allowlisted hosts only, with size and time limits. Every image is decoded and re-encoded (which strips anything hidden in the file) at the sizes the UI needs, stored by content hash, and served through the asset handler with long cache lifetimes. An accent colour is extracted per game for glows and the lightbar.
- User overrides (title, cover, hero) are never overwritten by a refresh.

## 6. Launching, tracking, controller

- **Launch:** Steam games use `steam://rungameid/…` (setting: direct exe). Everything else starts with `CreateProcess`: explicit path, working folder and argument array, no shell. Elevation happens only when the game's manifest requires it.
- **Tracking:** the process tree plus matching on the install folder, which covers launcher-to-game handoffs. It polls every 2 s only while a game runs. Playtime is saved every minute. Steam playtime is imported from `localconfig.vdf`. Optional passive tracking of games started outside WaterLauncher isn't built yet: it needs polling while idle, which works against the idle budget.
- **Command line:** `WaterLauncher.exe --play <id>` starts a game without opening the interface (for shortcuts).
- **Hooks pipeline:** before launch (Syncer sync, add-ons) and after exit (Syncer backup, add-ons). Each step has a timeout, progress and a skip option, and the result is shown in the launch sequence.
- **DualSense (SDL3):**
  - Features: hotplug, glyph detection (PS or Xbox), haptic ticks, a lightbar in the game's accent colour, and the PS button to summon the launcher.
  - While a game runs, WaterLauncher releases the controller and keeps only a passive PS-button listener. It sends no output reports and never switches the controller's report mode.
- **Per-game controller mode:** *Auto* / *Native* / *Steam Input*.
  - Auto uses Steam categories 55/57/58, Steam's controller support field, and whether `libScePad.dll` or SDL is in the game folder. A game goes through Steam Input only when a PlayStation controller is in use and the game supports Xbox controllers but not PlayStation ones.
  - The Steam Input route uses non-Steam shortcuts that WaterLauncher manages in `shortcuts.vdf`, in a "WaterLauncher" collection. Changes are applied while Steam is closed, or after asking to restart it. These games launch through `steam://rungameid/<shortcut>`.
- **Game mode:** all interface windows close; the smoke test measured about 420 MB for WebView2 alone. WaterLauncher stays in the tray (pulled forward from v1.0, because game mode needs it). Closing the window yourself quits, except while a game runs. Pressing PS opens a borderless topmost overlay window. Nothing is ever injected into games, so anti-cheat stays happy.
  - Measured in v0.4: the Go core uses about 35 MB while a game runs with the interface closed.
  - Passive listening restarts SDL without its HIDAPI drivers and with enhanced reports off, so nothing is written to the controller.

## 7. Interface

- The **Big picture** layouts share the focus engine (spatial navigation in zones), input intents (keyboard, or gamepad events from Go; the browser gamepad API isn't used), glyphs, sounds, haptics, an on-screen keyboard for search, and a controller-friendly Settings screen.
- **Desktop D:** filters, cover grid, details panel, Settings window. It auto-switches to big picture when a controller connects (setting).
- **Speed:** virtualised grids and rows, animations only on transform and opacity, images pre-sized, `content-visibility`, and reduced motion respected.
- **Themes:** CSS token sets per layout. User theme packs (tokens only, no JavaScript) come later.
- **Mock backend:** the frontend runs in a normal browser with fake data, for fast iteration and screenshots.

## 8. Syncer integration

Done in v0.5. Syncer's side is [ApolloF/syncer#7](https://github.com/ApolloF/syncer/pull/7); the protocol is Syncer's [docs/api.md](https://github.com/ApolloF/syncer/blob/main/docs/api.md), and WaterLauncher's use of it is in [syncer-api.md](syncer-api.md).

1. Syncer imports `gamekit` for VDF, Steam libraries, tamper checks and the manifest parser. `internal/steam` keeps its API as thin wrappers, and the manifest parser gives identical results (13,719 games with Windows saves).
2. `\\.\pipe\syncer`: JSON-RPC, one message per line, served by the window. A protected DACL grants only the current user, and remote clients are refused. `Syncer.exe --api` serves it without a window and exits a minute after the last connection. The client checks the server process's owner.
3. Methods: `status`, `games`, `gameStatus` (by title, Steam app, install folder, or a title the launcher registered), `syncNow` (with timeout), `backupNow` (optionally waiting), `conflicts`, `resolveConflict`, `open`, `registerGames`, `subscribe` (`changed` notifications).
4. The protocol is versioned (`status.protocol`, now 1). WaterLauncher handles Syncer being missing (*Get Syncer*) or older than 0.11 (*Update Syncer*; it never starts an old `Syncer.exe --api`, which would open its window).

Follow-ups in Syncer, after its `feature/steam-autocloud-copies` work lands: use `gamekit/steam.Accounts`, and let registered launcher games feed discovery (for now they only help `gameStatus`).

## 9. Add-ons

- **Manifest** `addon.json`: id, name, version, publisher, exe, protocol version, hooks, contributions (badges, actions, a settings schema) and permissions (for example `modifyGameFiles`, `network`).
- **Hooks:** `library.gameAdded`, `game.beforeLaunch` (progress, skip), `game.afterExit`, `game.status`, `game.actions`, `settings`.
- **Trust:** you enable each add-on explicitly and see its permissions. WaterLauncher pins its SHA-256 and asks again when it changes, and shows the Authenticode publisher when signed. Add-ons run with normal user rights, with timeouts and crash isolation.
- **DLSS Updater:** a new `--addon` mode, as a PR in `ApolloF/dlssupdater`. It needs the .NET 8 SDK, which gets installed in that phase.
  - Before launch: re-apply DLSS if a game patch reverted it.
  - Status: DLSS and OptiScaler versions.
  - Actions: install OptiScaler, restore DLSS, open in DLSS Updater.
- The full spec goes in `docs/addon-protocol.md`.

## 10. Security

- **WebView:** strict CSP with no remote scripts. Navigation away from the app is blocked, devtools and the context menu are off in release, only bound services are exposed, and every value from the UI is validated in Go (paths must resolve inside known roots).
- **Network:** allowlisted hosts, HTTPS, size limits, and images re-encoded. The only executable ever downloaded is WaterLauncher's own update from GitHub Releases, checked against its SHA-256.
- **Secrets:** the SteamGridDB key, Steam Web API key and Epic token are stored with DPAPI and never logged.
- **Parsers:** VDF, ini, JSON, lnk and ACF files are treated as untrusted, with size limits and fuzz tests.
- **Privileges:** WaterLauncher never elevates itself. The Syncer pipe is current-user only. Add-ons are covered in section 9.

## 11. Performance budget (this laptop: Ryzen 7 7840HS, 16 GB)

| Metric | Target |
|---|---|
| Cold start to interactive | < 1 s desktop, < 1.5 s big picture |
| Incremental scan / first full scan | < 300 ms / < 3 s |
| Navigation with 2,000 games | 60 fps |
| Idle CPU | about 0 % (event-driven, no polling without a running game) |
| Memory while playing | Go core < 50 MB (interface unloaded) |
| Installer | < 20 MB |

## 12. Phases

Each phase ends with a working build, a GitHub prerelease and a check-in.

| # | Version | Scope |
|---|---|---|
| 1 | v0.1 | Scaffold (Wails v3, Svelte 5, CI). Steam playtime import (pulled forward from v0.4). Detection code copied into `internal/` for now, split into `gamekit` in phase 5. SQLite schema, scanner (stores, unofficial, folders), Desktop D with real data (grid, filters, details), Settings skeleton, mock backend |
| 2 | v0.2 | Metadata and art (Steam, GOG, SteamGridDB; PCGamingWiki dropped), image pipeline, accent colours, *Found on this PC* review |
| 3 | v0.3 | Big picture: focus engine, SDL3 controller layer, **Deck** first, then Console, then Orbit; glyphs, haptics, lightbar, on-screen keyboard |
| 4 | v0.4 | Launch and tracking, hooks pipeline, game mode (with tray), PS-button overlay, Steam Input routing, `--play` (playtime import moved to v0.1) |
| 5 | v0.5 | `gamekit` repo and Syncer PR (pipe API, `--api`), launcher client, save status, conflicts, backup after exit |
| 6 | v0.6 | Add-on protocol, host and permissions UI; DLSS Updater `--addon` PR |
| 7 | v0.7 | Owned-but-not-installed games from Steam, GOG and Epic (opt-in) |
| 8 | v1.0 | NSIS installer, auto-update, start with Windows, code signing, docs, release |

## 13. Testing

- **Go:** fixture folders for every detection signature, golden tests for identification, fuzz tests for parsers, and tests for the launch and tracking state machine.
- **Frontend:** `svelte-check` and Vitest (focus engine, intents). The mock backend is checked visually at 1600×1000 and 1920×1080.
- **Performance:** a fake-library generator with 2,000 games, measuring fps and memory.
- **Controllers:** a checklist per phase covering DualSense over USB and Bluetooth, and an Xbox controller.
- **CI:** GitHub Actions on `windows-latest` runs the tests and a build. Tagged versions build the installer.

## 14. Risks

| Risk | Mitigation |
|---|---|
| Wails v3 is a beta that ships every few days | Pinned version, deliberate upgrades, thin Wails-specific layer |
| Steam only picks up `shortcuts.vdf` changes after a restart | Batch changes, apply when Steam is closed or after asking; only games that need Steam Input |
| DualSense shared between the launcher and a game (Bluetooth report mode) | Passive listener while playing, no output reports, can be turned off |
| Steam also claims the PS button | Setting explained in Controller settings; detect and warn |
| Overlay over exclusive full-screen games | Recommend borderless windowed; overlay is optional; no injection |
| Metadata endpoints change | One provider per package, cache, fallback chain |
| Scope | Only installed games are managed; no download sources or crack tools |

## 15. Workflow and releases

- `main` always builds. Each phase gets a branch (`feature/v0.1-foundation`, and so on) and a PR, merged after tests and the build pass. Small fixes can go straight to `main`.
- At the end of each phase, the tag `vX.Y.0` makes GitHub Actions build `WaterLauncher.exe` (plus the installer once it exists) with SHA-256 checksums, and publishes a **prerelease** with notes on what changed and what to test.
- The Syncer and DLSS Updater changes (phases 5 and 6) happen in their own repos: `gamekit` as a new public repo, and PRs (or direct pushes to `main`) for Syncer and DLSS Updater.
- License: MIT.

## 16. Environment (set up 2026-09-25)

- Go 1.27.0, Wails CLI v3.0.0-beta.26, Node 24.19 and npm 11.17, NSIS, Git 2.55, GitHub CLI 2.101 (signed in as ApolloF, used as the Git credential helper), WebView2 153.
- Repo: `C:\Users\Florian\Documents\Coding projects\WaterLauncher`, branch `main`, remote `https://github.com/ApolloF/WaterLauncher` (public). Repo-local identity `ApolloF <me@apollof.nl>`.
- CI (`.github/workflows/build.yml`) builds and tests every push and PR on `windows-latest`. A `v*` tag publishes a prerelease with the exe, its SHA-256, and `docs/releases/<tag>.md` as notes.
- Smoke test: a Wails v3 Svelte app built in 34 s into a 10.5 MB exe. At runtime the Go process used about 67 MB and WebView2 about 423 MB.
- Installed later: SDL3 3.4.16 (phase 3, from the libsdl-org release, hash checked) and the .NET 8 SDK (phase 6).
