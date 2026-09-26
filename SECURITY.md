# Security

## Reporting a problem

Please report security problems privately through GitHub: **Security → Report a vulnerability** on this repository. Don't open a public issue for them.

## What WaterLauncher does to stay safe

- **No elevation.** WaterLauncher, its installer and its updates run with your normal rights. A game that asks for administrator rights gets Windows' own prompt.
- **Nothing runs through a shell.** Games start with an explicit program path, working folder and arguments. Links handed to Windows are limited to store schemes (`steam://`, `com.epicgames.launcher://`, …) and `https://` pages.
- **Network.** Metadata, art and store accounts use allowlisted hosts over HTTPS, with size and time limits. Every image is decoded and re-encoded before it's stored, so only pixels reach the interface.
- **Updates** come only from this repository's GitHub releases (redirects are limited to GitHub's download hosts), must match the SHA-256 published with the release, and, once builds are signed, must carry the same publisher's Authenticode signature. See [docs/SIGNING.md](docs/SIGNING.md).
- **Secrets** (SteamGridDB key, Steam Web API key, Epic sign-in) are encrypted with Windows DPAPI for your account and never logged.
- **Add-ons** are off until you turn them on. Their program is pinned by SHA-256 and has to be approved again when it changes. They run as separate processes with your normal rights, with time limits.
- **Syncer's pipe** is checked to be served by a process of your own Windows account.
- **Interface.** A strict Content Security Policy (no remote code, no inline scripts), and every value from the interface is checked again in Go.
- **Parsers** for VDF, `.lnk`, SQLite and the rest treat files as untrusted, with size limits; several are fuzz-tested.
