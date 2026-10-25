# Testing Patterns

**Analysis Date:** 2026-10-25

## Test Framework

**Runner:**
- Go standard `testing`, with Go 1.26.0 declared in `go.mod` and installed by `.agents/setup`.
- Python scripts run contract and SDL UI tests with standard-library HTTP servers and subprocesses (`tests/http-contract.py`, `tests/ui-smoke.py`).
- Config: `scripts/check` defines the shared checks. `.github/workflows/check.yml` adds a separate UI test and device package job.

**Assertion Library:**
- Go uses explicit comparisons with `t.Fatal` and `t.Fatalf`. Python uses `assert`; no third-party assertion library is used in the inspected tests (`internal/library/library_test.go`, `tests/http-contract.py`).

**Run Commands:**
```bash
./scripts/check                  # Shared checks, race tests, build, HTTP contract
GOFLAGS='-tags=nodefaultfont,nonextuifonts' go test -race ./...
python3 tests/ui-smoke.py build/host/random-game-picker build/ui-smoke
GOFLAGS='-tags=nodefaultfont,nonextuifonts' go test -coverprofile=/tmp/picker-coverage.out ./...
```
- Run `.agents/setup` first to prepare Go, cgo, SDL libraries, ShellCheck, fonts, and Python/PyYAML. The UI test also needs `xvfb`, `xdotool`, and `imagemagick` (`.agents/setup`, `docs/development.md`).
- There is no watch command in `scripts/check` or `.github/workflows/check.yml`.
- The coverage command above is an optional local command, not a configured CI step. This map describes test contracts; a test list is not evidence of a passing run.

## Test File Organization

**Location:**
- Go tests are beside implementation files and use external test packages (`input_test`, `library_test`, and `picker_test`). See `internal/input/mapping_test.go`, `internal/library/library_test.go`, and `internal/picker/state_test.go`.
- Cross-process fixtures and their helper executable are under `tests/`.

**Naming:**
- Go files use `_test.go`; test functions use `Test` plus a behavior name (`internal/library/library_test.go`).
- Python suites use descriptive script names rather than a test-discovery framework (`tests/http-contract.py`, `tests/ui-smoke.py`).

**Structure:**
```text
internal/
  input/mapping_test.go
  library/library_test.go
  picker/state_test.go
tests/
  controller-events.cpp
  http-client/main.go
  http-contract.py
  ui-smoke.py
scripts/
  check
  check-docs.py
```

## Test Structure

**Suite Organization:**
```go
// internal/picker/state_test.go
func TestRevalidationRequiresSystemAndExactPath(t *testing.T) {
    game := library.Game{System: "snes", Path: "/games/a 'quoted'; $.rom"}
    for _, candidate := range []library.Game{{System: "other", Path: game.Path}, {System: game.System, Path: "/games/a.rom"}} {
        if picker.Revalidate(game, library.Library{Games: []library.Game{candidate}}) == nil {
            t.Fatal("Accepted a different launch identity")
        }
    }
    if err := picker.Revalidate(game, library.Library{Games: []library.Game{game}}); err != nil {
        t.Fatal(err)
    }
}
```

**Patterns:**
- Tests create isolated file trees with `t.TempDir()` and a local writer with `t.Helper()` (`internal/library/library_test.go`).
- State tests run a sequence of transitions and inspect public state. They also save and load settings through real temporary files (`internal/picker/state_test.go`).
- Tests stop on failed preconditions with `t.Fatal`; structured comparisons use `t.Fatalf` diagnostics (`internal/library/library_test.go`).
- Python suites close servers and join threads in `finally` blocks. The UI suite also terminates child processes and its X display (`tests/http-contract.py`, `tests/ui-smoke.py`).

## Mocking

**Framework:** No mock framework. Use injected functions, local HTTP servers, and a test-only SDL preload library (`internal/library/library_test.go`, `tests/http-contract.py`, `tests/controller-events.cpp`).

**Patterns:**
```go
// Route-based fake in internal/library/library_test.go
fetch := func(_ context.Context, route string) ([]byte, error) {
    switch route {
    case "/systems":
        return json.Marshal(systems)
    case "/systems/custom/games":
        return json.Marshal(games)
    case "/systems/ports/games":
        return json.Marshal(games[:1])
    default:
        return nil, fmt.Errorf("Unexpected route %s", route)
    }
}
```

**What to Mock:**
- Replace ES catalogue responses at the `library.Fetch` boundary for discovery tests (`internal/library/library_test.go`).
- Replace the ES service with a local HTTP server for transport and UI tests. Both suites bind fixed loopback port 1234; do not run them together or against a real ES service (`tests/http-contract.py`, `tests/ui-smoke.py`).
- Feed logical SDL press/release events through a FIFO and `LD_PRELOAD` for controller-layout checks (`tests/controller-events.cpp`, `tests/ui-smoke.py`).

**What NOT to Mock:**
- Existing tests use real file reads, XML/JSON parsing, filter logic, and seeded selection (`internal/library/library_test.go`).
- Transport tests use the real Go client executable and actual loopback requests (`tests/http-client/main.go`, `tests/http-contract.py`).
- UI tests use the actual host binary, SDL event queue, and window. They do not start an emulator or test a physical controller driver (`tests/ui-smoke.py`, `docs/development.md`).

## Fixtures and Factories

**Test Data:**
```go
// internal/picker/state_test.go
game := library.Game{System: "snes", Path: "/games/a 'quoted'; $.rom"}
```
- Discovery fixtures include relocated paths, shell-like characters, custom configuration priority, hidden and missing games, collections, duplicate paths, and malformed artwork XML (`internal/library/library_test.go`).
- UI fixtures generate synthetic ROM text, PNG artwork, XML configuration, and JSON catalogue responses in a temporary directory. The suite converts art to JPEG for the smaller-size run (`tests/ui-smoke.py`).
- Sampling uses a fixed PCG source rather than production random seeds (`internal/library/library_test.go`).

**Location:**
- Fixture definitions live in test source, not a shared `testdata/` directory (`internal/library/library_test.go`, `internal/picker/state_test.go`, `tests/ui-smoke.py`).
- UI captures go to the directory passed as the second argument. CI passes `build/ui-smoke/` and uploads that directory (`.github/workflows/check.yml`).

## Coverage

**Requirements:** None enforced by `scripts/check` or `.github/workflows/check.yml`. No coverage threshold, saved coverage report, or measured percentage is supplied here.

**View Coverage:**
```bash
GOFLAGS='-tags=nodefaultfont,nonextuifonts' go test -coverprofile=/tmp/picker-coverage.out ./...
go tool cover -func=/tmp/picker-coverage.out
go tool cover -html=/tmp/picker-coverage.out
```
- These commands measure Go test execution only. They do not measure the Python-driven host UI run, packaging assertions, or device behavior (`scripts/check`, `tests/ui-smoke.py`, `scripts/package-device`).
- Direct Go tests exist for input, library, picker and launcher. Firmware, UI and command paths are exercised by the real SDL/Python fixture rather than matching `_test.go` files (`tests/http-contract.py`, `tests/ui-smoke.py`, `internal/launcher/es_test.go`).
- Captures check that a rendered window has more than 16 colors. Artwork rejection also compares its pixels with the corrupt-art fallback. There are no golden UI baselines; inspect representative captures for clipping and readability (`tests/ui-smoke.py`).
- The map does not claim hardware, performance, or emulator success coverage. The full device matrix, MagicX behavior, startup budgets, and normal game return need separate evidence (`docs/testing.md`).
- Source/package version and release trigger match official main (`0.2.0` / `v0.2.0`). Local audit tests and builds do not replace or republish the stable archive (`cmd/random-game-picker/main.go`, `scripts/package-device`, `.github/workflows/release.yml`).

## Test Types

**Unit Tests:**
- Input checks KNULLI and Xbox-layout face buttons, ignored button releases, and keyboard Enter (`internal/input/mapping_test.go`).
- Picker checks saved startup filters, focus retention, recent exclusion and reset, settings round-trip, failure recovery, and exact system/path revalidation (`internal/picker/state_test.go`).
- Library checks discovery contracts, filter intersection, duplicate exclusion, metadata fallback, title-uniform sampling, and empty/singleton pools (`internal/library/library_test.go`).

**Integration Tests:**
- The HTTP contract checks exact Unicode/metacharacter body bytes, escaped routes, fixed loopback despite proxy settings, HTTP/framing/size failures, redirects, disconnect/truncation/timeout and one POST without retry. Go tests cancel an accepted but unacknowledged launch (`tests/http-contract.py`, `internal/launcher/es_test.go`).
- `scripts/check` also checks shell syntax, ShellCheck, documentation, action pins, Go formatting, module integrity, vet, race tests, and host builds. It compiles the SDL event shim but does not run the UI suite.
- The device workflow cross-builds on Debian Bookworm. Packaging checks AArch64 architecture, loader, dependency allowlist and GLIBC ceiling 2.34 with explicit failures. Controlled ELF fixtures reject higher versions even under `PYTHONOPTIMIZE=1`; they also verify archive bytes/digest and build provenance. This does not run the device binary (`scripts/build-device`, `scripts/package-device`, `tests/package-device.py`).
- `scripts/check-docs.py` checks `README.md`, `docs/*.md`, `.planning/codebase/*.md` and workflow files without contacting GitHub.
- New discovery fixtures cover symlink ambiguity/retargeting, conflicting additive configuration, cancellation and the XML size boundary. Settings fixtures cover whole-document fallback, exact/oversized reads, failed writes/renames and recovery (`internal/library/library_test.go`, `internal/picker/state_test.go`).
- SDL fixture layout checks now run both default KNULLI and Xbox at both target sizes. Artwork with invalid bytes, dimensions above 2048 or size above 8 MiB must render the same fallback. Both sizes still receive 100 rerolls (`tests/ui-smoke.py`).

**E2E Tests:**
- Host UI tests cover 1280×720 and 640×480, art and fallback, long titles, 100 rerolls per size, filters, empty pools, startup controls, delayed and failed scans, recovery, stale selections, and demo restrictions (`tests/ui-smoke.py`).
- The fake ES server checks that the UI window is gone before exactly one launch request arrives. This proves a host handoff boundary, not emulator execution or physical input compatibility (`tests/ui-smoke.py`).
- Manual device work must record firmware, hardware, controller, launch/return behavior, state preservation, and the tested version. The v0.2.0 release notes report limited TSP development-build evidence, including injected evdev input and a Gambatte handoff. Physical presses, normal hotkey return, and MagicX remain unverified (`docs/testing.md`).

## Common Patterns

**Async Testing:**
```python
# tests/ui-smoke.py: hold discovery while the user enters settings.
scan_allowed.clear()
# Start the real UI, wait for the first read, then open settings.
scan_allowed.set()
# Let discovery finish; Escape must still exit from settings.
```
- Use bounded waits for processes, HTTP calls, windows, and request counts. The UI test combines polling, events, and short sleeps; it is not a virtual-time test (`tests/ui-smoke.py`).
- Go race checks run only through `go test -race ./...`. The host binary used by the UI suite is built without `-race` in `scripts/check`.

**Error Testing:**
```go
// internal/library/library_test.go
if _, err := library.Choose(nil, rng); err == nil {
    t.Fatal("Empty pool accepted")
}
```
- Test failed discovery clears playable state and that a later scan recovers (`internal/picker/state_test.go`).
- Test a truncated response returns failure while the server sees only one submission for that request (`tests/http-contract.py`).
- Test removed games and unavailable catalogues prevent launch while leaving settings or rescan reachable (`tests/ui-smoke.py`).

---

*Testing analysis: 2026-10-25*
