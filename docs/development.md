# Development setup and CI

The source is the 0.2.1 maintenance release of the Go/gabagool rewrite. Host builds and fixture tests do not prove full handheld compatibility; see the [device test guide](testing.md) and [Go architecture](go-rewrite.md).

## Orb setup

Amp runs executable `.agents/setup` in a fresh orb after the file reaches the project's default branch. It installs missing Debian/Ubuntu packages for cgo, pkg-config, SDL2/image/ttf/gfx, Python/PyYAML, and ShellCheck, plus checksum-verified Go 1.26.0. It downloads the locked Go modules and skips package installation on warm runs. Run it manually with:

```sh
./.agents/setup
./scripts/check
```

No credentials, services, or environment-file changes are required for the host build. There is no resume hook because there is nothing to reconnect. Packages come from the host distribution; this is not a pinned KNULLI sysroot. The separate cross build uses Debian Bookworm and gates the actual ELF dependencies and GLIBC symbols.

## Checks

`scripts/lint` is the fast local/CI command. `scripts/check` runs it before the
race tests, host build and HTTP contracts. The checks include:

- Bash syntax and ShellCheck checks.
- Local Markdown file-link, whitespace, code-fence, and conflict-marker checks.
- Required PRD acceptance, open-question, and milestone marker checks.
- Workflow YAML parsing, full-commit action-pin checks and pinned actionlint v1.7.7.
- Go format, module integrity, vet and race-enabled tests for discovery, filtering, uniform selection, state and input contracts.
- A Go/cgo host build and real loopback HTTP transport with a fake ES server. The test-only C++ shim injects SDL events; the app has no C++ implementation.

Both prek and pre-commit read `.pre-commit-config.yaml`. Install the hooks with
`prek install` (or `pre-commit install --hook-type pre-commit --hook-type pre-push`).
The pre-commit stage runs `scripts/lint`; pre-push runs the Go race tests with
the same font build tags as CI. Run both stages before submitting changes:

```sh
prek run --all-files
prek run --all-files --hook-stage pre-push
```

The hooks require Go, cgo/SDL2/image/ttf/gfx, ShellCheck and Python/PyYAML.
The pinned actionlint tool is fetched through Go on first use. The complete
`scripts/check` and SDL smoke suite use the Linux setup above; CMake, CTest and
application C++ lint/build checks no longer apply. Keep the test-only SDL shim.

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

Keyboard controls: arrows, Enter (A), Escape (B), X (reroll), Y (settings from selector; rescan inside settings), F10 (quit). The demo also opens directly on a selected game, but never scans, launches or saves filters. `--es-config DIR` selects a custom ES user configuration directory for fixture testing. `PICKER_STATE_DIR` selects app-owned settings/log state. No CLI option accepts an emulator command. The Go binary prints `0.2.1`; other firmware profiles are rejected.

Build the AArch64 archive on Debian 12 with `./scripts/build-device`. This installs cross dependencies, builds Go with cgo/arm64, gates GLIBC at 2.34 or earlier, and calls `scripts/package-device`. SDL2_gfx is required in addition to SDL2/image/ttf. The ZIP, checksum, sizes, inventory, module licenses, ELF report and `build-info.txt` go under `build/release`. Build information records exact compiler/binutils, SDL, Debian package and Go module versions; it is provenance, not a pinned firmware sysroot or byte-reproducibility guarantee. A local rebuild is not automatically a published release or hardware compatibility proof.

The full check and cross build run `python3 tests/package-device.py`. It exercises
the real packager with controlled ELF reports, including GLIBC 2.35/2.36/2.40
rejection under normal and optimized Python. Production gates use explicit
failures, not assertions that `PYTHONOPTIMIZE` can remove. HTTP fixtures also
reject malformed framing, redirects and timeouts without retrying a POST.

### Performance audit

The selector redraws on SDL events, scan results and settings return, not every
idle poll. Text and eligible pools are rebuilt only for those redraws. Artwork
still decodes synchronously on the SDL thread, with one retained texture and the
existing byte/dimension bounds. No async SDL ownership or persistent cache was
added. Startup, rescan and Play still perform full discovery; path identities
are reused within each scan only. Symlink retargeting and cross-system ambiguity
are regression-tested. Discovery checks cancellation during game processing.

Run the host discovery benchmark with:

```sh
go test ./internal/library -run '^$' -bench BenchmarkDiscover -benchtime=3x -count=3
```

On this Debian 12 x86-64 orb, synthetic 10k/25k/50k catalogues with in-memory
HTTP responses had median scan times of 166→113 ms, 420→304 ms and 988→709 ms
before/after scan-local path reuse (three samples, three scans per sample).
At 50k, allocations fell from about 246 MB/2.75 million to 188 MB/2.00 million
per scan. These are not RSS, cold-start, real ES, SD-card or device latency
measurements; do not use them to sign off the plan's budgets.
Physical compatibility and precise performance remain tracked in issues
[#6](https://github.com/jellydn/knulli-random-game-picker/issues/6),
[#7](https://github.com/jellydn/knulli-random-game-picker/issues/7),
[#8](https://github.com/jellydn/knulli-random-game-picker/issues/8) and
[#9](https://github.com/jellydn/knulli-random-game-picker/issues/9).

The SDL suite covers both controller layouts at both sizes and compares
over-dimension/over-byte artwork with the corrupt-art fallback. To run the same
suite with race instrumentation, build the host binary with `go build -race`
and the normal font tags, then provide its adjacent assets and SDL shim. Neither
host variant starts an emulator or proves physical controls/audio/return.

Both build scripts use `-tags=nodefaultfont,nonextuifonts` to omit unused gabagool
fonts. The UI sets `FALLBACK_FONT` to the bundled DejaVu Sans before toolkit
initialization, so selector and settings share the same licensed font.

## GitHub Actions

`.github/workflows/check.yml` runs host and AArch64 package jobs on pull requests, pushes to `main`, and manual dispatch. It uses read-only repository permissions, disables persisted checkout credentials, limits runtime, and cancels obsolete checks on the same ref. It calls the same setup/build/check commands as local development. The host job also runs the headless SDL smoke test.

This follows the [KNULLI App Store setup](https://github.com/jellydn/knulli-app-store/blob/main/.agents/setup) and [check workflow](https://github.com/jellydn/knulli-app-store/blob/main/.github/workflows/check.yml) patterns. Its [original setup/CI thread](https://ampcode.com/threads/T-01a0a370-234b-729a-9b63-3871aa36a52c) records the method: inspect the environment, add repeatable setup, use shared checks, and verify locally before delivery. Action commit pins match the inspected App Store workflows; Go/catalogue commands do not belong in the picker build.

`.github/workflows/release.yml` publishes only the approved `v0.2.1` tag after
host/device jobs pass. The publication job verifies tag/version/asset identity
and checksums, then uploads the device job's exact package without rebuilding.
Only that job has write permission. It cannot publish under prior tags; earlier
release assets and tags stay unchanged. Future versions need a separately
reviewed version/tag and [App Store asset contract](go-rewrite.md#app-store-asset-contract).
