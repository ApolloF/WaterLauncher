# Code signing

WaterLauncher builds are unsigned until a code-signing certificate is set up. Everything works unsigned; the differences are:

- Windows SmartScreen warns on first run of a new download ("Windows protected your PC") until the file has built up reputation.
- The updater can only check a download against the SHA-256 published in the same GitHub release. Once builds are signed, it also requires every update to be signed by the same publisher as the running copy (`internal/update.CheckPublisher`), so a tampered release asset is refused even if its checksum file was replaced too.

## What's wired up

`build/windows/sign.ps1 <file>` signs a file with `signtool` (SHA-256, RFC 3161 timestamp) and verifies the result. With no certificate configured it prints a note and succeeds, so the same build runs everywhere. It reads:

| Variable | Meaning |
|---|---|
| `SIGN_PFX_BASE64`, `SIGN_PFX_PASSWORD` | A `.pfx` certificate, base64-encoded, and its password |
| `SIGN_THUMBPRINT` | Or: the SHA-1 thumbprint of a certificate in the current user's store (a hardware token, a cloud HSM's local provider) |
| `SIGN_TIMESTAMP_URL` | Timestamp server (default `http://timestamp.digicert.com`) |

CI (`.github/workflows/build.yml`) passes the repository secrets `SIGN_PFX_BASE64` and `SIGN_PFX_PASSWORD` through. When they're set it signs `WaterLauncher.exe`, the uninstaller (while NSIS builds it, through `-DSIGN_CMD`), and `WaterLauncher-setup.exe`. Checksums are made after signing.

To add the secrets:

```bash
base64 -w0 certificate.pfx | gh secret set SIGN_PFX_BASE64
gh secret set SIGN_PFX_PASSWORD
```

## Getting a certificate

Since June 2023, publicly trusted code-signing keys must live in hardware (a token or an HSM), so a plain `.pfx` from a CA is rare now. Options, cheapest first:

1. **[SignPath Foundation](https://signpath.org)**: free code signing for open-source projects. It signs in their HSM from a GitHub Actions step, after a review of the project. Its action would replace the `sign.ps1` calls in CI.
2. **Azure Artifact Signing** (formerly Trusted Signing): a monthly fee, identity validation for an individual or organization, signing from CI through `signtool` with Microsoft's dlib. `sign.ps1` would pass `/dlib` and `/dmdf` instead of `/f`.
3. **An OV certificate on a USB token** from a CA: works with `SIGN_THUMBPRINT` for local release builds; not usable from GitHub's hosted runners.

Whichever it is, keep the publisher name stable: the updater compares the signer's name, and an update signed under a different name is refused. When the name has to change, ship one release that's installed by hand.
