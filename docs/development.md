# Development setup and CI

The published release is [v0.2.0](https://github.com/jellydn/knulli-random-game-picker/releases/tag/v0.2.0),
the Go/gabagool rewrite. This checkout still uses `0.2.0-dev` in the binary and
package script, with a historical `v0.1.1` release trigger. Do not use that trigger
to republish or replace assets. Host builds and fixture tests do not prove KNULLI
compatibility; see the [device test guide](testing.md) and [Go architecture](go-rewrite.md).

## Orb setup

Amp runs executable `.agents/setup` in a fresh orb after the file reaches the project's default branch. It installs missing Debian/Ubuntu packages for cgo, pkg-config, SDL2/image/ttf/gfx, Python/PyYAML, and ShellCheck, plus checksum-verified Go 1.26.0. It downloads the locked Go modules and skips package installation on warm runs. Run it manually with:

```sh
./.agents/setup
./scripts/check
```

No credentials, services, or environment-file changes are required for the host build. There is no resume hook because there is nothing to reconnect. Packages come from the host distribution; this is not a pinned KNULLI sysroot. The separate cross build uses Debian Bookworm and gates the actual ELF dependencies and GLIBC symbols.

## Checks

`scripts/check` is the shared local/CI command. It performs:

- Bash syntax and ShellCheck checks.
- Local Markdown file-link, whitespace, code-fence, and conflict-marker checks.
- Required PRD acceptance, open-question, and milestone marker checks.
- Workflow YAML parsing and full-commit action-pin checks. This is not a full GitHub Actions schema validator.
- Go format, module integrity, vet and race-enabled tests for discovery, filtering, uniform selection, state and input contracts.
- A Go/cgo host build and real loopback HTTP transport with a fake ES server. The test-only C++ shim injects SDL events; the app has no C++ implementation.

For the SDL smoke test, install `xvfb`, `xdotool`, and `imagemagick`, then run:

```sh
python3 tests/ui-smoke.py build/host/random-game-picker build/ui-smoke
```

It starts a temporary X display and fake ES server, exercises both sizes and 100 rerolls per size, and verifies that the window closes before one exact-path launch request. A test-only preload library feeds logical SDL presses/releases through the event queue: both default KNULLI and Xbox-layout modes must play with the first A after startup, reroll with X without a scan, open settings with Y, rescan with Y inside settings, return/quit with B, and quit with Start. Saved filters, empty pools, delayed/failed discovery and recovery remain covered. It does not simulate a physical controller driver, start an emulator or prove hardware mappings. It needs loopback port 1234 free. Captures use synthetic games and artwork.

The suite also blocks stale selections, rejects unknown firmware, and verifies
that demo mode does not scan, submit or save filters. It captures long-title and
missing-art states. With `ffmpeg` installed, set `PICKER_UI_RECORD` to an absolute
`.webm` path to record the 640×480 selector-to-settings keyboard interaction.

For a desktop demo (launch and persistent writes disabled):

```sh
./build/host/random-game-picker --demo --windowed --size 640x480
```

Keyboard controls: arrows, Enter (A), Escape (B), X (reroll), Y (settings from selector; rescan inside settings), F10 (quit). The demo also opens directly on a selected game, but never scans, launches or saves filters. `--es-config DIR` selects a custom ES user configuration directory for fixture testing. `PICKER_STATE_DIR` selects app-owned settings/log state. No CLI option accepts an emulator command. The Go binary prints `0.2.0-dev`; other firmware profiles are rejected.

Build an experimental AArch64 archive on Debian 12 with `./scripts/build-device`. This installs cross dependencies, builds Go with cgo/arm64, inspects dynamic dependencies and GLIBC symbols, and calls `scripts/package-device`. SDL2_gfx is now required in addition to SDL2/image/ttf. The dev ZIP, checksum, sizes, inventory, module licenses and ELF report go under `build/release`. It is not a published release or hardware compatibility proof.

Both build scripts use `-tags=nodefaultfont,nonextuifonts` to omit unused gabagool
fonts. The UI sets `FALLBACK_FONT` to the bundled DejaVu Sans before toolkit
initialization, so selector and settings share the same licensed font.

## GitHub Actions

`.github/workflows/check.yml` runs host and AArch64 package jobs on pull requests, pushes to `main`, and manual dispatch. It uses read-only repository permissions, disables persisted checkout credentials, limits runtime, and cancels obsolete checks on the same ref. It calls the same setup/build/check commands as local development. The host job also runs the headless SDL smoke test.

This follows the [KNULLI App Store setup](https://github.com/jellydn/knulli-app-store/blob/main/.agents/setup) and [check workflow](https://github.com/jellydn/knulli-app-store/blob/main/.github/workflows/check.yml) patterns. Its [original setup/CI thread](https://ampcode.com/threads/T-01a0a370-234b-729a-9b63-3871aa36a52c) records the method: inspect the environment, add repeatable setup, use shared checks, and verify locally before delivery. Action commit pins match the inspected App Store workflows; Go/catalogue commands do not belong in the picker build.

In this checkout, `.github/workflows/release.yml` triggers only for `v0.1.1`,
after host and device jobs pass. The publication job uploads the device job's
exact bytes without rebuilding. This historical workflow does not describe the
published v0.2.0 release process. Do not replace published assets or move a
published tag. A future release needs a reviewed version/trigger change.
Stable release status does not close the full M0 hardware gates.
