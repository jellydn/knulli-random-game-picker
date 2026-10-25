# Codebase Structure

**Analysis Date:** 2026-10-25

## Directory Layout

```text
knulli-random-game-picker/
├── .agents/setup                 # Host toolchain and native dependencies
├── .github/workflows/            # Host/device checks and gated v0.2.0 release
├── .planning/codebase/           # Codebase map documents
├── cmd/random-game-picker/       # Application main package
├── internal/
│   ├── firmware/                 # KNULLI policy and paths
│   ├── input/                    # SDL logical input maps and tests
│   ├── launcher/                 # Local ES HTTP transport
│   ├── library/                  # Discovery, eligibility, sampling, and tests
│   ├── picker/                   # Typed state, persistence, and tests
│   └── ui/                       # Gabagool shell, settings, and SDL canvas
├── packaging/                    # Installed KNULLI Port wrapper
├── scripts/                      # Checks, cross-build, and package assembly
├── tests/
│   ├── http-client/main.go       # Test executable for real transport calls
│   ├── http-contract.py          # Fake server and HTTP assertions
│   ├── ui-smoke.py               # Real SDL UI with fake ES and Xvfb
│   ├── package-device.py         # Real packager with controlled ELF reports
│   └── controller-events.cpp     # Test-only logical SDL event injection
├── docs/                         # Requirements, design, setup, and device checks
├── go.mod                        # Module name, Go version, direct dependencies
├── go.sum                        # Module checksums
├── README.md                     # User and contributor quick start
├── LICENSE                       # Project license
├── THIRD_PARTY.md                # Dependency notices
├── renovate.json                 # Dependency update configuration
└── .gitignore                    # Build, Python cache, and portal exclusions
```

## Directory Purposes

**`cmd/random-game-picker/`:**
- Purpose: Own process startup and final handoff.
- Contains: One Go `main` package. It parses flags, locks the OS thread, loads/saves filters, supplies the scan callback, and submits after UI shutdown.
- Key files: `cmd/random-game-picker/main.go`.

**`internal/firmware/`:**
- Purpose: Own firmware paths and launch/input policy.
- Contains: A concrete KNULLI `Profile` resolver. Unknown firmware names fail.
- Key files: `internal/firmware/profile.go`.

**`internal/input/`:**
- Purpose: Own SDL logical button and keyboard translation.
- Contains: Semantic actions, Gabagool virtual-button maps, layout correction, and tests for press/release behavior.
- Key files: `internal/input/mapping.go`, `internal/input/mapping_test.go`.

**`internal/launcher/`:**
- Purpose: Own loopback ES transport, not emulator commands.
- Contains: Request bounds, HTTP response checks, GET fetch, and one exact-path POST adapter.
- Key files: `internal/launcher/es.go`, `internal/launcher/es_test.go`. Additional transport integration tests are in `tests/http-contract.py` and `tests/http-client/main.go`.

**`internal/library/`:**
- Purpose: Own read-only catalogue construction and selection inputs.
- Contains: ES configuration/XML discovery, live JSON reconciliation, metadata paths, duplicate checks, eligibility, and uniform sampling.
- Key files: `internal/library/library.go`, `internal/library/library_test.go`.

**`internal/picker/`:**
- Purpose: Own selection state without SDL or HTTP.
- Contains: View/operation enums, filters, focus, recent choices, revalidation, and JSON settings load/save.
- Key files: `internal/picker/state.go`, `internal/picker/state_test.go`.

**`internal/ui/`:**
- Purpose: Own the display and input resource lifetime.
- Contains: `Run`, background scan channels, settings callbacks, the custom selector canvas, font/artwork resources, and PNG capture.
- Key files: `internal/ui/run.go`, `internal/ui/canvas.go`.

**`tests/`:**
- Purpose: Check boundaries that need real processes, sockets, or SDL.
- Contains: A Go transport client, Python fake ES servers, Xvfb UI scenarios, and a C++ preload shim. The shim is test code; the application implementation is Go.
- Key files: `tests/http-client/main.go`, `tests/http-contract.py`, `tests/ui-smoke.py`, `tests/controller-events.cpp`.

**`packaging/`:**
- Purpose: Supply the installed KNULLI Port entry.
- Contains: A shell wrapper that resolves its own directory, sets package-local state and layout policy, and executes the adjacent binary.
- Key files: `packaging/RandomGamePicker.sh`.

**`scripts/`:**
- Purpose: Supply shared local and CI commands.
- Contains: Shell/Go checks, Markdown/workflow checks, Debian 12 cross-build, archive assembly, ABI validation, inventory, and checksums.
- Key files: `scripts/check`, `scripts/check-docs.py`, `scripts/build-device`, `scripts/package-device`.

**`docs/`:**
- Purpose: Keep requirements, design records, setup instructions, and hardware test scope.
- Contains: PRD baseline, implementation milestones, Go rewrite boundaries, development instructions, and installation/device checks.
- Key files: `docs/prd.md`, `docs/implementation-plan.md`, `docs/go-rewrite.md`, `docs/development.md`, `docs/testing.md`.

**`.agents/` and `.github/workflows/`:**
- Purpose: Prepare the host and run automated checks or publication jobs.
- Contains: Executable setup plus host/device check and gated `v0.2.0` release workflows.
- Key files: `.agents/setup`, `.github/workflows/check.yml`, `.github/workflows/release.yml`.

## Key File Locations

**Entry Points:**
- `cmd/random-game-picker/main.go`: Application startup and handoff after SDL teardown.
- `packaging/RandomGamePicker.sh`: Installed ES Port launcher.
- `scripts/check`: Shared host verification command. The SDL smoke command runs separately in CI.
- `scripts/build-device`: Debian Bookworm arm64 cross-build; calls package assembly.
- `tests/http-client/main.go`: Test-only transport executable, not a second application.

**Configuration:**
- `go.mod`, `go.sum`: Go 1.26.0 module definition and locked dependency checksums.
- `internal/firmware/profile.go`: Default ES configuration/state paths, controller DB, and environment overrides.
- `cmd/random-game-picker/main.go`: CLI flags such as `--assets`, `--demo`, `--windowed`, `--size`, and `--es-config`.
- `packaging/RandomGamePicker.sh`: `PICKER_STATE_DIR` and layout selection from the Ports Xbox-layout flag.
- `.github/workflows/check.yml`: Pull-request, main-push, and manual host/device checks.
- `.github/workflows/release.yml`: Trigger restricted to `v0.2.0`, with tag/version/asset/checksum validation before publication.
- `scripts/package-device`: Archive version `0.2.0`, GLIBC ceiling 2.34 and ABI/package rules; copies the build's `build-info.txt` provenance sidecar.
- `scripts/lint` and `.pre-commit-config.yaml`: Shared local/CI checks and lint/race hook stages.
- `tests/package-device.py`: Real packager tests with controlled ELF reports; `internal/launcher/es_test.go` tests live-loopback cancellation and invalid requests.
- `renovate.json`: Dependency update configuration.

**Core Logic:**
- `internal/library/library.go`: `Discover`, `Eligible`, and `Choose`.
- `internal/picker/state.go`: `State`, view transitions, recent choices, `Revalidate`, `Load`, and `Save`.
- `internal/launcher/es.go`: `Request`, `Fetch`, and `Submit`.
- `internal/input/mapping.go`: `Action`, `Mapping`, and view-specific virtual-button maps.
- `internal/ui/run.go`: UI loop, settings list, async scan handling, and launch verification.
- `internal/ui/canvas.go`: Selector layout, artwork checks, texture lifetime, and screenshot output.

**Testing:**
- `internal/library/library_test.go`: Discovery precedence, live metadata, filtering, duplicate exclusion, and title-uniform sampling.
- `internal/picker/state_test.go`: Saved-filter behavior, focus retention, recent exclusion, persistence, recovery, and exact launch identity.
- `internal/input/mapping_test.go`: KNULLI/Xbox face-button semantics and ignored release events.
- `tests/http-contract.py`: Real loopback requests, status failures, exact POST bodies, and no retry after a truncated response.
- `tests/ui-smoke.py`: Real host UI at both target sizes, layout input, saved filters, empty/stale libraries, failed/delayed discovery, demo isolation, and window closure before launch.
- `tests/controller-events.cpp`: Logical SDL button press/release injection for host UI checks, not a physical-controller emulator.
- `scripts/check-docs.py`: Checks `README.md`, `docs/*.md`, requirement markers, workflow YAML, and pinned actions. It does not include `.planning/codebase/` in its document list.
- `docs/testing.md`: Manual device checks. Host fixtures do not establish handheld compatibility.

## Naming Conventions

**Files:**
- Go implementation files use short lower-case names: `state.go`, `mapping.go`, `canvas.go` (`internal/picker/`, `internal/input/`, `internal/ui/`).
- Go tests use `_test.go` beside their owner: `internal/library/library_test.go`.
- Command packages use `main.go`: `cmd/random-game-picker/main.go`, `tests/http-client/main.go`.
- Executable shell tools use extensionless lower-case names: `scripts/check`, `scripts/build-device`, `scripts/package-device`.
- Python/C++ boundary tests use descriptive hyphenated names: `tests/ui-smoke.py`, `tests/controller-events.cpp`.
- The installed Port entry retains its display-oriented name: `packaging/RandomGamePicker.sh`.

**Directories:**
- Internal package names describe one responsibility: `library`, `picker`, `launcher`, `firmware`, `input`, and `ui` (`internal/`).
- Executable directories use the command name: `cmd/random-game-picker/` and `tests/http-client/`.
- Build directories describe target or output purpose: `build/host/`, `build/aarch64/`, `build/release/`, and `build/ui-smoke/` (`scripts/check`, `scripts/build-device`, `.github/workflows/check.yml`).

## Where to Add New Code

**New Feature:**
- Primary code: Use `internal/library/library.go` for catalogue/eligibility changes, `internal/picker/state.go` for state rules, and `internal/ui/` for display/actions. Keep final launch wiring in `cmd/random-game-picker/main.go`.
- Tests: Add owner-level cases beside the package. Use `tests/http-contract.py` for transport behavior and `tests/ui-smoke.py` for cross-package UI behavior.

**New Component/Module:**
- Implementation: Place a distinct internal responsibility under `internal/` and connect it in `cmd/random-game-picker/main.go`. Use existing function boundaries `library.Fetch` and `ui.Scanner` when they fit.
- Implementation: Firmware policy belongs in `internal/firmware/`; input policy belongs in `internal/input/`. A new firmware name also needs validated discovery/launch, package, and device contracts, not only a path record (`docs/go-rewrite.md`).

**Utilities:**
- Shared helpers: Keep path and metadata helpers in `internal/library/library.go`, drawing/resource helpers in `internal/ui/canvas.go`, and request checks in `internal/launcher/es.go`. No general utility package exists.

## Special Directories

**`build/`:**
- Purpose: Generated host/device binaries, copied fonts, test shim, UI evidence, release ZIP, licenses, and ABI/checksum/inventory reports (`scripts/check`, `scripts/build-device`, `scripts/package-device`, `tests/ui-smoke.py`).
- Generated: Yes, by build and test commands.
- Committed: No; excluded by `.gitignore`.

**Installed `assets/` and `licenses/`:**
- Purpose: Runtime DejaVu Sans and license notices in the staged device archive. These are package outputs, not source asset directories (`scripts/package-device`).
- Generated: Yes, copied during build/package assembly.
- Committed: No generated copies; original project notices are `LICENSE` and `THIRD_PARTY.md`.

**Installed `state/`:**
- Purpose: Package-local `settings.json`, its temporary replacement, `last-run.log`, and toolkit `ui.log`. Direct binary execution instead uses the profile default or environment override (`packaging/RandomGamePicker.sh`, `internal/picker/state.go`, `internal/firmware/profile.go`, `internal/ui/run.go`).
- Generated: Yes, at runtime. No persistent library cache is written.
- Committed: Not present in the source tree. A repository-root `state/` is not explicitly ignored by `.gitignore`.

**`.planning/codebase/`:**
- Purpose: Source map refreshed after alignment with official GitHub main `da79d67`, including local audit changes. It is not a map of the public archive's contents.
- Generated: Yes, from source inspection and local verification. `CONCERNS.md` records resolved findings and open hardware issues #6–#9.
- Committed: Tracked project documentation; `.gitignore` does not exclude it. `scripts/check-docs.py` checks its hygiene and local links.

**`.amp/portals/` and Python `__pycache__/`:**
- Purpose: Orb portal metadata and Python bytecode caches, respectively (`.gitignore`).
- Generated: Yes, by the corresponding tools when used.
- Committed: No; excluded by `.gitignore`.

---

*Structure analysis: 2026-10-25*
