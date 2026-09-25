# SDL3

`SDL3.dll` is the official Windows x64 build of [SDL](https://github.com/libsdl-org/SDL) 3.4.16, from `SDL3-3.4.16-win32-x64.zip` on the project's GitHub release `release-3.4.16`. It is used only for game controller input, rumble and the DualSense lightbar.

- SHA-256: `1f98969319302a100931f4385e5918a0bd53ab07773040682d22e7edb54858c0` (checked by `internal/pad` before every load)
- License: zlib, see [LICENSE.txt](LICENSE.txt)

To update: download the new `win32-x64` zip from the release, replace the DLL, and update the hash here and in `internal/pad/dll_windows.go`.
