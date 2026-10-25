# Architecture

**Analysis Date:** 2026-10-25

## Pattern Overview

**Overall:** One Go desktop process with separate UI, selection, discovery, and local transport packages.

**Key Characteristics:**
- The command connects concrete packages. There is no service container or general router (`cmd/random-game-picker/main.go`).
- The UI uses a typed state model. Background scans return data through a channel. The UI thread applies each result (`internal/picker/state.go`, `internal/ui/run.go`).
- ES owns the running game catalogue and emulator launch. The picker reads existing files and sends one exact-path request after SDL closes (`internal/library/library.go`, `internal/launcher/es.go`).
- This map describes the 0.2.1 maintenance changes after alignment with official GitHub main `da79d67`. Binary/package version is `0.2.1`; release trigger is `v0.2.1`. Prior releases and assets are preserved (`cmd/random-game-picker/main.go`, `scripts/package-device`, `.github/workflows/release.yml`).
- GitHub [v0.2.0](https://github.com/jellydn/knulli-random-game-picker/releases/tag/v0.2.0) is published. Its archive checksum and actual ELF dependency/GLIBC report were checked, but the published binary was not run on hardware here.
- The selector redraws after events and scan results, not on every idle poll. Artwork decoding remains synchronous on the SDL thread. Discovery reuses path identities within each scan only; Play still performs full fresh discovery (`internal/ui/run.go`, `internal/ui/canvas.go`, `internal/library/library.go`).

## Layers

**Command and Process Lifetime:**
- Purpose: Configure the process and connect the packages.
- Location: `cmd/random-game-picker/main.go`
- Contains: Flags, OS-thread lock, profile selection, filter load/save, scan callback, and final launch submission.
- Depends on: `internal/firmware`, `internal/library`, `internal/picker`, `internal/ui`, `internal/launcher`.
- Used by: The installed wrapper and direct desktop invocation (`packaging/RandomGamePicker.sh`, `docs/development.md`).

**Firmware Policy:**
- Purpose: Supply device paths, controller policy, and launch capability.
- Location: `internal/firmware/profile.go`
- Contains: `Profile` and KNULLI-only `Resolve`.
- Depends on: `internal/library` for `Locations`, and environment variables for home, state, and layout.
- Used by: The command and UI (`cmd/random-game-picker/main.go`, `internal/ui/run.go`).

**UI and Input:**
- Purpose: Show the selector and settings. Convert logical SDL input to actions.
- Location: `internal/ui/`, `internal/input/`
- Contains: Gabagool lifetime and settings list, custom SDL canvas, artwork checks, screenshots, and semantic input maps.
- Depends on: Gabagool/v2, go-sdl2, firmware policy, library data, and picker state (`go.mod`, `internal/ui/run.go`, `internal/ui/canvas.go`, `internal/input/mapping.go`).
- Used by: The command. UI code returns a game; it does not submit a launch (`cmd/random-game-picker/main.go`).

**Selection and Session State:**
- Purpose: Keep view transitions, filters, selection, and recent choices separate from SDL.
- Location: `internal/picker/state.go`
- Contains: `State`, `Filters`, view and operation enums, settings actions, persistence, and `Revalidate`.
- Depends on: `internal/library`, standard JSON/file packages, and `math/rand/v2`.
- Used by: The command, UI, and state tests (`cmd/random-game-picker/main.go`, `internal/ui/run.go`, `internal/picker/state_test.go`).

**Catalogue and Eligibility:**
- Purpose: Build a read-only game list from ES configuration, live records, and artwork metadata.
- Location: `internal/library/library.go`
- Contains: XML/JSON readers, path resolution, duplicate exclusion, `Eligible`, and uniform `Choose`.
- Depends on: Standard packages and an injected `Fetch` function. It does not import the launcher or SDL.
- Used by: The command's scan callback, picker state, and library tests (`cmd/random-game-picker/main.go`, `internal/picker/state.go`, `internal/library/library_test.go`).

**ES Transport:**
- Purpose: Read the local ES API and submit a launch path.
- Location: `internal/launcher/es.go`
- Contains: Bounded `Request`, GET adapter `Fetch`, and POST adapter `Submit`.
- Depends on: Standard TCP, HTTP, context, and time packages. The fixed endpoint is `127.0.0.1:1234`.
- Used by: The command and the transport test client (`cmd/random-game-picker/main.go`, `tests/http-client/main.go`).

**Device Wrapper and Delivery:**
- Purpose: Set Port policy and assemble a device package without replacing firmware libraries.
- Location: `packaging/RandomGamePicker.sh`, `scripts/build-device`, `scripts/package-device`
- Contains: State/log setup, controller-layout flag handling, Debian 12 AArch64 cgo build, ABI checks, ZIP inventory, checksums, and licenses.
- Depends on: The built Go binary, system SDL libraries, DejaVu Sans, and module license files.
- Used by: KNULLI Ports and workflow device jobs (`.github/workflows/check.yml`, `.github/workflows/release.yml`).

## Data Flow

**Startup and Discovery:**
1. The wrapper sets package-local state and KNULLI/Xbox layout policy. It replaces itself with the binary and redirects output to `state/last-run.log` (`packaging/RandomGamePicker.sh`).
2. The command locks the OS thread, parses flags, rejects unsupported profiles or sizes, and loads saved filters outside demo mode (`cmd/random-game-picker/main.go`).
3. The UI initializes Gabagool and SDL, then starts `Scanner` in a goroutine. A buffered channel carries the result back to the UI thread (`internal/ui/run.go`).
4. Discovery selects the first existing base configuration: user custom, user standard, shared standard, then legacy standard. It merges named fragments from user and shared directories in that order (`internal/library/library.go`).
5. Discovery fetches `/systems`, then game records for each real non-collection system. Configured roots determine support; live records determine membership, favorites, hidden status, and launch paths (`internal/library/library.go`).
6. Discovery reads optional gamelist metadata for titles/artwork. It excludes Ports, missing or hidden games, the picker wrapper, and ambiguous canonical paths. Unsupported real systems still contribute to duplicate counts (`internal/library/library.go`).
7. `State.Accept` applies the library, retains a valid saved system filter, and rolls if the selector is active and the pool is not empty. It does not move focus away from settings (`internal/picker/state.go`).

**Reroll and Settings:**
1. `Mapping` converts keyboard or logical controller events to actions. The selector uses the same semantic actions for footer labels (`internal/input/mapping.go`, `internal/ui/canvas.go`).
2. Reroll calls `State.Roll` on the in-memory eligible pool; it does not scan ES. Each eligible title has equal probability (`internal/picker/state.go`, `internal/library/library.go`).
3. Back or Settings opens the Gabagool list. Left/right use list cycle callbacks only in that view. System, favorites, recent exclusion, and reset update picker state (`internal/ui/run.go`, `internal/picker/state.go`).
4. Settings rescan starts discovery only when idle. The UI applies its result through `tick` (`internal/ui/run.go`).

**Play and Exit:**
1. Confirm copies the selected game and starts a new full scan. Demo mode and `CanLaunch=false` block this action (`internal/ui/run.go`).
2. `Revalidate` requires the same system and exact path in the fresh library. Failure clears launch readiness and shows a notice (`internal/picker/state.go`, `internal/ui/run.go`).
3. Success exits the loop. Deferred cleanup closes canvas resources, controllers, and Gabagool before `Run` returns the game (`internal/ui/run.go`, `internal/ui/canvas.go`).
4. The command cancels the scan context and saves filters. A save error prevents submission (`cmd/random-game-picker/main.go`).
5. The command submits one `POST /launch` with the exact path as `text/plain`, using a separate background context. No emulator command or automatic retry exists (`cmd/random-game-picker/main.go`, `internal/launcher/es.go`).
6. HTTP success proves API acceptance only. Real emulator start, return, saves, and hardware input require separate device checks (`docs/testing.md`).

**Build and Verification:**
1. `.agents/setup` prepares the Go/cgo host environment. `scripts/check` calls shared `scripts/lint`, then package-policy tests, race tests, the host build and loopback contracts. `.pre-commit-config.yaml` shares lint/race checks with CI.
2. `tests/ui-smoke.py` runs the real host UI with a fake ES server and synthetic files under Xvfb. `tests/controller-events.cpp` injects logical SDL events, not hardware events.
3. `scripts/build-device` cross-builds for Linux arm64 on Debian 12 and records exact inputs in `build-info.txt`. `scripts/package-device` checks the ELF library allowlist and GLIBC ceiling of 2.34 with explicit non-assertion failures, then writes the archive and reports.
4. `.github/workflows/check.yml` runs host and device jobs. `.github/workflows/release.yml` permits only `v0.2.1`, verifies identity/checksums, and uploads the tested device bytes without rebuilding. The approved tag starts publication only after both gates pass.

**State Management:**
- `State` owns the in-memory library, selected index, view, operation, focus, notice, random generator, and last ten picks (`internal/picker/state.go`).
- Only `Filters.System` and `Filters.Favorites` persist. `Save` writes a temporary JSON file and renames it. Recent exclusion and recent paths are session-only (`internal/picker/state.go`).
- There is no persistent library index. Startup, manual rescan, and Play verification each read ES again (`internal/ui/run.go`).
- Demo mode supplies two synthetic games and skips discovery, filter load/save, and launch. Toolkit logging can still write a temporary log (`cmd/random-game-picker/main.go`, `internal/ui/run.go`).

## Key Abstractions

**Game and Library:**
- Purpose: Carry launch identity, display metadata, favorites, and warnings.
- Examples: `internal/library/library.go`, `internal/library/library_test.go`
- Pattern: Plain structs; the original live path remains the launch path, while canonical paths detect duplicates.

**Fetch and Scanner:**
- Purpose: Separate discovery from transport and asynchronous UI work.
- Examples: `internal/library/library.go`, `internal/ui/run.go`, `cmd/random-game-picker/main.go`
- Pattern: Function injection. Tests can supply records without changing the discovery code.

**State, View, and Operation:**
- Purpose: Separate user view, scan progress, and selection rules.
- Examples: `internal/picker/state.go`, `internal/picker/state_test.go`
- Pattern: Typed enums and explicit methods, without SDL or HTTP dependencies.

**Profile and Mapping:**
- Purpose: Keep firmware paths/capabilities separate from logical action mapping.
- Examples: `internal/firmware/profile.go`, `internal/input/mapping.go`, `internal/input/mapping_test.go`
- Pattern: One supported profile plus generated keyboard/controller maps. This is not a multi-firmware backend registry.

## Entry Points

**Application Binary:**
- Location: `cmd/random-game-picker/main.go`
- Triggers: Direct execution or the Port wrapper.
- Responsibilities: Startup, demo/version modes, package wiring, persistence, and submission after UI cleanup.

**KNULLI Port Wrapper:**
- Location: `packaging/RandomGamePicker.sh`
- Triggers: ES Ports launch.
- Responsibilities: Resolve package location, select layout policy, create state directory, replace the session log, and execute the binary.

**Checks and Device Packaging:**
- Location: `scripts/check`, `scripts/build-device`, `scripts/package-device`
- Triggers: Developer commands and GitHub workflows.
- Responsibilities: Verify host contracts, build the device binary, and inspect/package the resulting bytes.

## Error Handling

**Strategy:** Return fatal errors to the command; keep recoverable discovery and selection errors visible in UI state (`cmd/random-game-picker/main.go`, `internal/picker/state.go`).

**Patterns:**
- Missing system configuration, invalid system XML, or failed live API reads fail discovery. Invalid optional gamelist XML produces a warning instead (`internal/library/library.go`).
- Empty selection pools return an error. Recent exclusion is not silently relaxed (`internal/library/library.go`, `internal/picker/state.go`).
- Missing, large, or unsupported artwork uses a visual fallback. Some render errors skip that draw operation (`internal/ui/canvas.go`).
- Invalid or unreadable saved filters leave default values. Save errors propagate and stop exit-time launch submission (`internal/picker/state.go`, `cmd/random-game-picker/main.go`).
- Transport rejects non-2xx status, unsupported framing, oversized bodies, and incomplete responses. Some interrupted requests have an unknown launch outcome; callers must not retry automatically (`internal/launcher/es.go`).

## Cross-Cutting Concerns

**Logging:** The command prints firmware and handoff messages. The wrapper replaces `state/last-run.log` each session. Gabagool uses `ui.log` in the profile state directory, or a temporary demo log (`cmd/random-game-picker/main.go`, `packaging/RandomGamePicker.sh`, `internal/ui/run.go`).

**Validation:** Configuration reads are capped at 32 MiB. Live paths must be bounded, non-NUL, absolute, present, and unambiguous to enter the library. Artwork must be PNG/JPEG, at most 8 MiB, and at most 2048 pixels on each axis. Transport has five-second bounds and a 64 MiB body cap (`internal/library/library.go`, `internal/ui/canvas.go`, `internal/launcher/es.go`).

**Authentication:** No account or token flow exists. The client connects directly to fixed IPv4 loopback without DNS, proxy, redirects, or retries. The local ES service is the trust boundary, not an authenticated remote provider (`internal/launcher/es.go`).

---

*Architecture analysis: 2026-10-25*
