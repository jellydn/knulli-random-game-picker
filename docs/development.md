# Development setup and CI

The repository contains the first experimental application. Host builds and fixture tests do not prove KNULLI compatibility; see the [device test guide](testing.md).

## Orb setup

Amp runs executable `.agents/setup` in a fresh orb after the file reaches the project's default branch. It installs missing Debian/Ubuntu packages for C++, CMake, Ninja, pkg-config, SDL2/image/ttf, pugixml, nlohmann/json, Python/PyYAML, and ShellCheck. It skips package installation on warm runs. Run it manually with:

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
- A temporary C++17 compile/link/load probe for the selected host dependencies. It opens no window and uses no controller or network. Temporary files are removed on exit.
- CMake/Ninja application build and CTest contracts for library discovery, filtering, uniform selection, and real loopback HTTP transport with a fake ES server.

For the SDL smoke test, install `xvfb`, `xdotool`, and `imagemagick`, then run:

```sh
python3 tests/ui-smoke.py build/host/random-game-picker build/ui-smoke
```

It starts a temporary X display and fake ES server, exercises both sizes and 100 rerolls per size, and verifies that the window closes before one exact-path launch request. It never starts an emulator. It needs loopback port 1234 free. Captures use a synthetic game and artwork, not a real device library.

For a desktop demo (launch and persistent writes disabled):

```sh
./build/host/random-game-picker --demo --windowed --size 640x480
```

Keyboard controls: arrows, Enter (A), Escape (B), X (reroll), Y (rescan). `--es-config DIR` selects a custom ES user configuration directory for fixture testing. `PICKER_STATE_DIR` selects app-owned settings/log state. No CLI option accepts an emulator command.

Build an experimental AArch64 archive on Debian 12 with `./scripts/build-device`. This installs cross dependencies, builds with static pugixml/libstdc++/libgcc, inspects dynamic dependencies and GLIBC symbols, and calls `scripts/package-device`. The ZIP, checksum, exact sizes, inventory, and ELF report go under `build/release`. Archive entry ordering and timestamps are stable; toolchain/package versions can still change binary bytes, so manifests must use the exact published asset, not a later rebuild.

## GitHub Actions

`.github/workflows/check.yml` runs host and AArch64 package jobs on pull requests, pushes to `main`, and manual dispatch. It uses read-only repository permissions, disables persisted checkout credentials, limits runtime, and cancels obsolete checks on the same ref. It calls the same setup/build/check commands as local development. The host job also runs the headless SDL smoke test.

This follows the [KNULLI App Store setup](https://github.com/jellydn/knulli-app-store/blob/main/.agents/setup) and [check workflow](https://github.com/jellydn/knulli-app-store/blob/main/.github/workflows/check.yml) patterns. Its [original setup/CI thread](https://ampcode.com/threads/T-01a0a370-234b-729a-9b63-3871aa36a52c) records the method: inspect the environment, add repeatable setup, use shared checks, and verify locally before delivery. Action commit pins match the inspected App Store workflows; Go/catalogue commands do not belong in the picker build.

The owner approved the public GitHub mirror and first experimental release. `.github/workflows/release.yml` publishes only the `v0.1.0-alpha.1` tag, after host and device jobs pass; the publication job uploads the device job's exact bytes without rebuilding. Do not replace published assets or move a published tag. A later version needs a reviewed version/trigger change. Publishing the alpha does not close M0 hardware gates.
