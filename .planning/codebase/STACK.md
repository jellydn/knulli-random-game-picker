# Technology Stack

**Analysis Date:** 2026-10-25

## Languages

**Primary:**
- Go 1.26.0 - Application entry point, library discovery, selection, firmware policy, input, HTTP transport, and UI. Module: `github.com/jellydn/knulli-random-game-picker`. Evidence: `go.mod`, `cmd/random-game-picker/main.go`, `internal/`.

**Secondary:**
- Bash and POSIX shell - Host setup, checks, cross-build, package creation, and the Ports launcher. Evidence: `.agents/setup`, `scripts/check`, `scripts/build-device`, `scripts/package-device`, `packaging/RandomGamePicker.sh`.
- Python 3 - Document checks, HTTP fixtures, UI smoke tests, and ZIP creation. Evidence: `scripts/check-docs.py`, `tests/http-contract.py`, `tests/ui-smoke.py`, `scripts/package-device`.
- C++ - Test-only SDL event injection library. The current application is not C++. Evidence: `tests/controller-events.cpp`, `scripts/check`.
- YAML - GitHub Actions jobs. Evidence: `.github/workflows/check.yml`, `.github/workflows/release.yml`.

## Runtime

**Environment:**
- Native Linux application with Go and cgo. SDL needs the main OS thread; the entry point calls `runtime.LockOSThread()`. Evidence: `cmd/random-game-picker/main.go`, `scripts/build-device`.
- KNULLI is the only supported firmware. Other profile names produce an error. Evidence: `internal/firmware/profile.go`.
- Source aligned with official GitHub main `da79d67` reports `0.2.1`; package version is also `0.2.1`. This maintenance release preserves prior public packages. Evidence: `cmd/random-game-picker/main.go`, `scripts/package-device`.
- Published version `v0.2.0` has asset `random-game-picker-0.2.0-knulli-aarch64.zip`. The audit checked its checksum and actual ELF, without executing it on hardware. Evidence: [release record](https://github.com/jellydn/knulli-random-game-picker/releases/tag/v0.2.0).

**Package Manager:**
- Go modules with Go 1.26.0. Dependency versions are in `go.mod`; integrity hashes are in `go.sum`. Evidence: `go.mod`, `.agents/setup`.
- Lockfile: `go.sum` is present. `scripts/check` runs `go mod verify`.
- Debian/Ubuntu APT supplies native development libraries and tools. Exact package versions are not pinned. Evidence: `.agents/setup`, `scripts/build-device`.

## Frameworks

**Core:**
- gabagool/v2 v2.26.6 - Console UI setup, settings list, input mapping, and toolkit lifecycle. Evidence: `go.mod`, `internal/ui/run.go`, `internal/input/mapping.go`.
- go-sdl2 v0.4.40 - SDL window, controller, texture, image, and font bindings. Evidence: `go.mod`, `internal/ui/run.go`, `internal/ui/canvas.go`.

**Testing:**
- Go standard `testing` package - Library, picker state, and input unit tests. Race checks use `go test -race ./...`. Evidence: `internal/library/library_test.go`, `internal/picker/state_test.go`, `internal/input/mapping_test.go`, `scripts/check`.
- Python standard library - Fake ES HTTP service and application smoke tests. Evidence: `tests/http-contract.py`, `tests/ui-smoke.py`.
- Xvfb, xdotool, and ImageMagick - Headless display, input, and UI evidence. Optional ffmpeg records UI interaction. Evidence: `.github/workflows/check.yml`, `tests/ui-smoke.py`, `docs/development.md`.

**Build/Dev:**
- Go compiler, gofmt, vet, module verification, and race detector - Shared checks and host builds. Evidence: `scripts/check`.
- GCC AArch64 cross-compiler, GNU strip/readelf, and pkg-config - Linux arm64 cgo build and ELF inspection. Evidence: `scripts/build-device`, `scripts/package-device`.
- ShellCheck and PyYAML - Shell checks and workflow parsing. Evidence: `.agents/setup`, `scripts/check`, `scripts/check-docs.py`.
- Renovate recommended preset - Dependency update configuration. Evidence: `renovate.json`.

## Key Dependencies

**Critical:**
- `github.com/BrandonKowalski/gabagool/v2` v2.26.6 and `github.com/veandco/go-sdl2` v0.4.40 are the two direct Go dependencies. Evidence: `go.mod`.
- Firmware supplies SDL2, SDL2_image, SDL2_ttf, SDL2_gfx, and glibc. The ZIP does not bundle or replace these libraries. Native versions are not pinned in source. Evidence: `.agents/setup`, `scripts/build-device`, `scripts/package-device`, `THIRD_PARTY.md`.
- DejaVu Sans - Bundled font used by both selector and settings. Build scripts copy it from the host package. Evidence: `scripts/check`, `scripts/build-device`, `internal/ui/run.go`, `internal/ui/canvas.go`.

**Infrastructure:**
- Indirect modules: certifiable v1.3.0, BurntSushi/toml v1.6.0, go-evdev `v0.0.0-20250804134636-ab1d56a1fe83`, go-i18n/v2 v2.6.1, oksvg `v0.0.0-20221011165216-be6e8873101c`, and rasterx `v0.0.0-20220730225603-2ab79fcdd4ef`. These are transitive dependencies, not separate application integrations. Evidence: `go.mod`.
- Other indirect modules: go.uber.org/atomic v1.11.0, golang.org/x/image v0.34.0, golang.org/x/net v0.48.0, and golang.org/x/text v0.33.0. Evidence: `go.mod`.
- Standard Go XML, JSON, filesystem, random, and network packages implement the application data and IPC contracts. No database driver or cloud SDK is declared. Evidence: `go.mod`, `internal/library/library.go`, `internal/picker/state.go`, `internal/launcher/es.go`.

## Configuration

**Environment:**
- No `.env` file or application credential is required. `PICKER_STATE_DIR` overrides the state location. The Ports launcher sets it to its own `state/` directory. Direct execution defaults to `/userdata/system/configs/random-game-picker`. Evidence: `packaging/RandomGamePicker.sh`, `internal/firmware/profile.go`, `docs/development.md`.
- `PICKER_CONTROLLER_LAYOUT=knulli` enables the KNULLI face-button correction. The wrapper selects `xbox` if `/var/run/ports_xbox_layout.flag` exists. SDL loads `/tmp/gamecontrollerdb.txt`. Evidence: `packaging/RandomGamePicker.sh`, `internal/firmware/profile.go`, `internal/ui/run.go`.
- CLI options include `--assets`, `--demo`, `--windowed`, `--size`, `--es-config`, `--screenshot`, `--version`, and `--firmware`. Demo mode does not scan, submit launches, or save filters. Evidence: `cmd/random-game-picker/main.go`, `internal/ui/run.go`.
- UI setup sets `FALLBACK_FONT`; windowed mode sets `ENVIRONMENT`, `WINDOW_WIDTH`, and `WINDOW_HEIGHT`. It clears `FLIP_FACE_BUTTONS` to prevent a second button swap. Evidence: `internal/ui/run.go`.

**Build:**
- `.agents/setup` installs host tools and checksum-verified Go 1.26.0, then downloads modules. It supports a Linux amd64 host. Evidence: `.agents/setup`.
- `scripts/check` builds the host binary. `scripts/build-device` requires Debian 12 and sets `CGO_ENABLED=1`, `GOOS=linux`, `GOARCH=arm64`, `CC=aarch64-linux-gnu-gcc`, and arm64 pkg-config paths.
- Both builds set `GOFLAGS=-tags=nodefaultfont,nonextuifonts` to omit unused toolkit fonts. Evidence: `scripts/check`, `scripts/build-device`.
- `scripts/package-device` creates a flat ZIP, SHA-256 file, inventory, `artifact.json`, licenses, and `abi.txt` under `build/release/`. It checks the ELF loader, allowed shared libraries, and GLIBC symbol versions.
- CI uses Ubuntu 24.04 host runners and a Debian Bookworm device-build container. Action references use full commit hashes. Evidence: `.github/workflows/check.yml`, `.github/workflows/release.yml`.

## Platform Requirements

**Development:**
- Debian/Ubuntu Linux amd64 with Go 1.26.0, a C/C++ toolchain, pkg-config, SDL2/image/ttf/gfx development packages, Python 3, PyYAML, ShellCheck, and DejaVu fonts. Evidence: `.agents/setup`.
- UI tests also need Xvfb, xdotool, ImageMagick, and a free loopback TCP port 1234. Evidence: `.github/workflows/check.yml`, `tests/ui-smoke.py`, `tests/http-contract.py`.
- Cross-build needs Debian 12, arm64 development packages, and an AArch64 GCC toolchain. Evidence: `scripts/build-device`.

**Production:**
- KNULLI Linux AArch64 only. Target devices are TrimUI Smart Pro at 1280×720 and MagicX Zero 28 at 640×480. These targets do not imply complete hardware validation. Evidence: `README.md`, `docs/testing.md`, `internal/firmware/profile.go`.
- The published `v0.2.0` binary requires glibc 2.34+ and SDL2/image/ttf/gfx. Its release notes describe limited TSP tests and list MagicX and full physical gameplay/return checks as unverified. Evidence: [published release notes](https://github.com/jellydn/knulli-random-game-picker/releases/tag/v0.2.0).
- The package check permits a maximum required GLIBC symbol version of 2.34, matching the published floor. Explicit failures cannot be disabled by Python optimization. The actual ELF-derived floor is written to `artifact.json`; exact toolchain/package/module inputs go to `build-info.txt`. Evidence: `scripts/package-device`, `scripts/build-device`, `tests/package-device.py`.
- ES must expose its local HTTP API during the Port session. Install the complete archive inside the configured Ports root and preserve `state/` on update. Evidence: `internal/launcher/es.go`, `packaging/RandomGamePicker.sh`, `docs/testing.md`.
- Release workflow is pinned to the approved `v0.2.1` tag and verifies tag/version/asset/checksums before publication. Shared `scripts/lint` includes pinned actionlint; `.pre-commit-config.yaml` supplies lint/race hooks. Evidence: `.github/workflows/release.yml`, `scripts/lint`, `docs/development.md`.

---

*Stack analysis: 2026-10-25*
