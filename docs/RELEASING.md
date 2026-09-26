# Releasing

From v1.1, every release is signed with WaterLauncher's **release key**, an ed25519 key that never goes to GitHub. The updater (v1.1 and later) installs only files whose SHA-256 is listed in a `SHA256SUMS` signed with that key for that exact tag. So even someone who took over the GitHub account, or replaced release files, can't ship an update.

## The key

- It lives on the maintainer's PC, encrypted with Windows DPAPI for that Windows account: `%APPDATA%\WaterLauncher-release\release-key.dpapi` (outside WaterLauncher's own data folder, so uninstalling the app can't delete it).
- Its public half is in `internal/update/keys.go`.
- **Backup** (do this once, and keep it offline, apart from its password):

  ```bash
  go run ./tools/release backup E:\waterlauncher-release-key.json
  ```

  The file is the key encrypted with AES-256-GCM under your password (PBKDF2-SHA256, 600,000 rounds). Without a backup, a lost PC means v1.1+ users can't receive updates until they install a new version by hand.
- On a new PC: `go run ./tools/release restore <file>`.
- Rotating: add the new public key to `keys.go` next to the old one, release that version (signed with the old key), then sign with the new key and remove the old one a release later.

## Steps

1. Merge the release branch into `main`; commit "Version X.Y.Z" (the version in `build/windows/info.json`, the status line in `docs/PLAN.md`); write `docs/releases/vX.Y.Z.md`.
2. Tag and push: `git tag vX.Y.Z && git push origin main vX.Y.Z`.
3. CI builds, tests and (when configured) code-signs, then creates a **draft** release with `WaterLauncher-setup.exe`, `WaterLauncher.exe` and their `.sha256`. Drafts are invisible to users and to the updater.
4. Sign and publish from the maintainer's PC:

   ```bash
   go run ./tools/release publish vX.Y.Z
   ```

   It downloads the draft's files, checks them against the `.sha256` CI made, prints the hashes, writes `SHA256SUMS` and `SHA256SUMS.sig`, uploads them, publishes the release (as *latest* unless it's a preview like `v1.2.0-beta.1`), and verifies the published signature.
5. Optional: compare the printed hashes with the ones in the CI run's *Checksums* step log before step 4. The signature says "the maintainer approved exactly these files"; it can't tell whether CI built the right thing.

`go run ./tools/release verify vX.Y.Z` checks any published release.
