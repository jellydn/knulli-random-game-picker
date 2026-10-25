# Coding Conventions

**Analysis Date:** 2026-10-25

## Naming Patterns

**Files:**
- Use short, lower-case Go file names by function: `internal/library/library.go`, `internal/picker/state.go`, and `internal/ui/canvas.go`.
- Put Go tests in matching `_test.go` files, for example `internal/picker/state_test.go`.
- Use lower-case names with hyphens for support scripts: `scripts/check-docs.py`, `scripts/build-device`, and `tests/ui-smoke.py`.

**Functions:**
- Use mixed case. Exported functions start with a capital letter: `Discover`, `Eligible`, and `Choose` in `internal/library/library.go`.
- Keep implementation functions private with a lower-case first letter: `readLimited` in `internal/library/library.go` and `showSettings` in `internal/ui/run.go`.
- Use `Test` plus a behavior name for Go tests, such as `TestRevalidationRequiresSystemAndExactPath` in `internal/picker/state_test.go`.

**Variables:**
- Use short local names where context is clear: `ctx`, `err`, `lib`, and `rng` in `internal/library/library.go`.
- Use descriptive names for state and policy: `SwapFaceButtons` in `internal/firmware/profile.go` and `ExcludeRecent` in `internal/picker/state.go`.
- Use upper-case environment names with underscores, such as `PICKER_STATE_DIR` and `PICKER_CONTROLLER_LAYOUT` in `packaging/RandomGamePicker.sh`.

**Types:**
- Use named structs for data boundaries: `library.Game`, `library.Locations`, `picker.Filters`, and `ui.Options` in their `internal/` packages.
- Use integer types with `iota` for actions, views, and operations in `internal/input/mapping.go` and `internal/picker/state.go`.
- Use function types for injected work: `library.Fetch` in `internal/library/library.go` and `ui.Scanner` in `internal/ui/run.go`.
- Use JSON and XML tags for external field names. Examples include `json:"favorites"` in `internal/picker/state.go` and `xml:"fullname"` in `internal/library/library.go`.

## Code Style

**Formatting:**
- `scripts/check` requires no output from `gofmt -l cmd internal tests/http-client`. Go files use tabs and standard Go formatting.
- `scripts/check` uses the build tags `nodefaultfont,nonextuifonts`. Use these tags for equivalent local Go checks and builds.
- Shell scripts use strict mode: Bash uses `set -euo pipefail` in `scripts/check`; POSIX shell uses `set -eu` in `packaging/RandomGamePicker.sh`.
- No Python or C++ formatter command is present in `scripts/check`. Python uses four-space indentation in `scripts/check-docs.py` and `tests/http-contract.py`.

**Linting:**
- `scripts/check` runs Bash and POSIX shell syntax checks, ShellCheck, `go mod verify`, and `go vet ./...`.
- ShellCheck exclusions have a local explanation. Examples cover test-shim compiler flags in `scripts/check` and the OS-release read in `scripts/build-device`.
- `scripts/check-docs.py` checks local links, final newlines, trailing spaces, code fences, conflict markers, required plan markers, workflow YAML, and full-commit action pins.
- The document checker reads `README.md`, `docs/*.md`, and `.planning/codebase/*.md`. Shared lint also runs pinned actionlint for GitHub Actions schema and expression checks (`scripts/check-docs.py`, `scripts/lint`, `docs/development.md`).

## Import Organization

**Order:**
1. Use a parenthesized Go import block. Standard-library imports form the first group in `cmd/random-game-picker/main.go` and `internal/ui/run.go`.
2. A blank line separates project or dependency imports in those files. Other files mix standard-library and module imports in one group, for example `internal/picker/state.go`. There is no uniform group rule beyond the format check in `scripts/check`.
3. Register image decoders with blank imports in `internal/ui/canvas.go`. The test shim uses system includes in `tests/controller-events.cpp`.

**Path Aliases:**
- No Go path-alias configuration is present in `go.mod`. Internal imports use the full module path `github.com/jellydn/knulli-random-game-picker/internal/...`.
- The UI uses the local import alias `gaba` for gabagool in `internal/ui/run.go` and `internal/ui/canvas.go`.

## Error Handling

**Patterns:**
- Return `error` to the caller and stop at failed prerequisites. `cmd/random-game-picker/main.go` prints the final error to standard error and exits with status 1.
- Add context with `fmt.Errorf` and `%w` where useful, as in `internal/launcher/es.go` and `internal/library/library.go`.
- Separate recoverable metadata warnings from fatal discovery errors. Bad artwork XML adds a warning; invalid system configuration stops discovery (`internal/library/library.go`).
- Set explicit unavailable state after a failed scan. Clear the selected game and show the error in `internal/picker/state.go` and `internal/ui/run.go`.
- Use `defer` for files, connections, SDL objects, and cancellation. Examples are in `internal/launcher/es.go`, `internal/ui/canvas.go`, and `cmd/random-game-picker/main.go`.
- Some optional operations discard errors explicitly. Examples include settings decode in `internal/picker/state.go` and renderer calls in `internal/ui/canvas.go`. Do not describe this as a universal error-propagation rule.
- Do not retry an uncertain launch submission. The transport reports an unknown outcome after interrupted or incomplete responses (`internal/launcher/es.go`).

## Logging

**Framework:** Standard Go `fmt` output and gabagool file diagnostics (`cmd/random-game-picker/main.go`, `internal/ui/run.go`).

**Patterns:**
- Print the firmware and button policy at startup. Print the handoff boundary before submission (`cmd/random-game-picker/main.go`).
- The Port wrapper replaces `state/last-run.log` each session and redirects both standard streams there (`packaging/RandomGamePicker.sh`).
- Toolkit output goes to `ui.log` under the profile state directory. Demo mode uses a temporary-directory log (`internal/ui/run.go`).
- A successful HTTP response means acceptance, not proof that a game ran. Keep that distinction in messages and evidence (`cmd/random-game-picker/main.go`, `docs/testing.md`).

## Comments

**When to Comment:**
- Use package comments to state ownership and safety boundaries in `internal/firmware/profile.go`, `internal/input/mapping.go`, `internal/library/library.go`, and `internal/picker/state.go`.
- Explain non-obvious constraints. Examples cover response routing in `internal/launcher/es.go`, duplicate-path exclusion in `internal/library/library.go`, and button-map ownership in `internal/ui/run.go`.
- Test comments explain which regression or safety property a sequence checks (`tests/http-contract.py`, `tests/ui-smoke.py`).

**JSDoc/TSDoc:**
- Not used in the inspected Go, Python, shell, and C++ sources. Go declaration comments appear on selected public functions, such as `Request` in `internal/launcher/es.go` and `Run` in `internal/ui/run.go`; public declarations are not all documented.

## Function Design

**Size:** No function-length limit is enforced by `scripts/check`. Small helpers coexist with longer orchestration functions, including `Discover` in `internal/library/library.go` and `Run` in `internal/ui/run.go`.

**Parameters:** Pass `context.Context` to discovery, UI, and transport work. Pass configuration structs and function dependencies instead of reading firmware paths in the picker (`internal/library/library.go`, `internal/ui/run.go`, `internal/picker/state.go`).

**Return Values:** Use `(value, error)` for fallible operations. `ui.Run` returns a revalidated game only after UI cleanup; the caller submits it. `picker.Load` instead returns default filters when loading fails (`internal/ui/run.go`, `cmd/random-game-picker/main.go`, `internal/picker/state.go`).

## Module Design

**Shared maintenance gates:** `scripts/lint` runs formatting, vet, module checks, shell/document/workflow lint and pinned actionlint. `scripts/check` adds packager-policy tests, race tests, host/shim builds and HTTP contracts. `.pre-commit-config.yaml` shares lint/race stages with CI. Keep codemap links and limits current; `scripts/check-docs.py` now checks these map files too.

**Safe optimization boundary:** Keep full fresh discovery before Play. Path identity reuse ends with one scan. Selector idle redraws are event-driven; SDL textures remain owned by the UI thread. Test symlink changes, both layouts/sizes and stale selections when touching these paths (`internal/library/library.go`, `internal/ui/run.go`, `tests/ui-smoke.py`).

**Exports:** Keep application packages under `internal/`. The entry point connects firmware policy, discovery, state, UI, and launch submission. Export data types and functions needed at those boundaries (`cmd/random-game-picker/main.go`, `internal/firmware/profile.go`, `internal/library/library.go`).

**Barrel Files:** Not used. Go files declare their package directly. There is no central re-export module in the inspected `cmd/` or `internal/` sources.

---

*Convention analysis: 2026-10-25*
