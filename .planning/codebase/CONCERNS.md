# Codebase Concerns

**Analysis Date:** 2026-10-25

## Tech Debt

**Stale-checkout findings resolved by upstream alignment:**
- The audit started at Amp checkout `ef63feb`. Official GitHub main was `da79d67`; it was merged before implementation. The seven refreshed map files were retained.
- Official main already sets binary/package version `0.2.0`, release trigger `v0.2.0`, and GLIBC ceiling **2.34**, not 2.36. Those old concerns were real in the obsolete snapshot, not in official main (`cmd/random-game-picker/main.go`, `scripts/package-device`, `.github/workflows/release.yml`).
- The public ZIP was downloaded and its SHA-256 matched `3f8a225cffcb2a89ffee166ff4f06dbf87923142309d97203e738a8200737df4`. Its actual ELF requires GLIBC 2.34 and SDL2/image/ttf/gfx, matching `artifact.json`. No published tag or asset was changed.
- The packager now uses explicit failures rather than Python assertions, so `PYTHONOPTIMIZE=1` cannot bypass the ABI gate. Regression cases reject 2.35, 2.36, 2.40, missing symbol data, wrong architecture and unexpected libraries (`tests/package-device.py`).
- Cross builds now record exact toolchain, target SDL, Debian package and Go module versions in `build-info.txt`. Distribution inputs remain unpinned; this is provenance, not byte-reproducibility or a target-firmware sysroot (`scripts/build-device`).

## Known Bugs

**Corrupt settings could activate partial filters — fixed:**
- A valid `system` followed by a wrong-typed `favorites` field caused JSON decoding to fail after changing the system filter. Loading now returns defaults for the whole invalid document and caps the actual read at 32 KiB (`internal/picker/state.go`).
- Regression tests cover the failing mixed-type case, malformed/oversized settings, the exact size boundary, failed temporary writes/renames, unchanged old bytes and save recovery (`internal/picker/state_test.go`). Save failure still prevents launch.

**Reported emulator exit is not a confirmed picker defect:**
- The [v0.2.0 release notes](https://github.com/jellydn/knulli-random-game-picker/releases/tag/v0.2.0) report Game Boy/Gambatte startup on TSP followed by RetroAchievements login errors and ES display restoration. This is not proof of sustained play or a normal hotkey return. No account/emulator workaround was added.

## Security Considerations

**Fixed local ES remains a trusted dependency:**
- `internal/launcher/es.go` uses fixed IPv4 loopback port 1234 without DNS, proxies, redirects, shell commands or retries. Loopback is not authentication; firmware and other local processes remain trusted.
- `tests/http-contract.py` exercises exact non-ASCII/metacharacter/newline body bytes, escaped routes, unusable proxies, malformed/conflicting lengths, oversized/chunked/unframed responses, redirects, disconnects, truncation and accepted-but-unacknowledged POST timeout. `internal/launcher/es_test.go` covers cancellation after acceptance without resubmission.

**Local file bounds are retained, not concurrent-replacement protection:**
- XML is capped at 32 MiB; HTTP bodies at 64 MiB; paths/titles at 4096 bytes; settings at 32 KiB; PNG/JPEG artwork at 8 MiB and 2048×2048 (`internal/library/library.go`, `internal/launcher/es.go`, `internal/picker/state.go`, `internal/ui/canvas.go`).
- Public discovery fixtures cover XML at/over its limit. SDL fixtures compare corrupt/over-dimension/over-byte artwork with the fallback. The private wrapper umask and atomic settings rename are unchanged (`packaging/RandomGamePicker.sh`).
- Files can still change between validation and use. Artwork is checked and then reopened by SDL. No hostile concurrent replacement, filesystem durability after power loss, or firmware authentication claim is made. Share redacted logs only.

## Performance Bottlenecks

**Full discovery remains intentional before Play:**
- Startup, explicit rescan and Play verification still read configuration, live systems/games and gamelists. No persistent cache was added. A cached record must never authorize launch (`internal/library/library.go`, `internal/ui/run.go`).
- Path identities are now reused only within one scan; the next scan resolves them again. Symlink aliases across unsupported real systems remain ambiguous. Regression tests change a symlink between scans to detect stale identity reuse. Cancellation is checked while processing games.
- `BenchmarkDiscover` measures 10k/25k/50k synthetic local files with in-memory HTTP bodies. Three samples of three scans each gave median scan times of 166→113 ms, 420→304 ms and 988→709 ms. At 50k, allocations fell from about 246 MB/2.75 million to 188 MB/2.00 million per scan (`internal/library/library_test.go`). This excludes real ES, SD-card and artwork latency and is not an RSS or device budget result.
- Device startup, tail latency and large-library limits remain owned by [#9](https://github.com/jellydn/knulli-random-game-picker/issues/9). Cache invalidation needs measured justification and a separate design; the PRD cache budgets remain unmet/unverified, not silently relaxed.

**Repeated idle rendering removed; synchronous artwork remains bounded:**
- `internal/ui/run.go` redraws the selector after SDL events, scan completion or return from settings, rather than every 16 ms. Unchanged idle text surfaces and eligible pools are no longer rebuilt. Window expose/resize events still redraw. The event loop still polls with a 16 ms sleep; no fixed FPS is claimed.
- `internal/ui/canvas.go` still decodes selected artwork synchronously and retains one texture. This preserves SDL-thread ownership and the released UI. Gabagool owns settings rendering. Async artwork, text caches and pool caches were not added without device evidence; measure them under #9 before increasing complexity.

## Fragile Areas

**Launch and Port exit:**
- Preserve full fresh discovery, exact system/path match, all SDL/input cleanup, successful settings save, then one POST. HTTP 2xx means submitted, not played (`cmd/random-game-picker/main.go`, `internal/picker/state.go`, `internal/ui/run.go`, `internal/launcher/es.go`).
- Stale selections, delayed/failed scans, save failures and ambiguous transport outcomes are host-covered. Inspect logs before manual retry. Real ES queue, override and return behavior remain external checks, not host-test conclusions.

**Controller and configuration contracts:**
- The wrapper's Xbox flag, firmware profile and semantic action map remain unchanged. Logical SDL tests now exercise both default KNULLI and Xbox layouts at both 1280×720 and 640×480 (`tests/ui-smoke.py`, `tests/controller-events.cpp`, `internal/input/mapping.go`).
- Configuration tests preserve base priority and sorted user-then-shared additive merging with nonempty field overrides. Shared additions currently win conflicts; this is the source contract, not proof of firmware-wide precedence. Do not change it without firmware evidence (`internal/library/library_test.go`).

## Scaling Limits

- Discovery remains sequential per live system, keeps games in memory and fails closed on any system request error. Five seconds is a per-connection deadline, not a whole-scan deadline (`internal/library/library.go`, `internal/launcher/es.go`).
- Recent history retains ten paths in session memory. Exclusion can empty a small pool and never silently relaxes. No universal maximum catalogue size, memory limit, FPS, or device latency is established (`internal/picker/state.go`, `docs/implementation-plan.md`).

## Dependencies at Risk

- Firmware must provide SDL2/image/ttf/gfx, AArch64 loader and glibc 2.34+. Libraries are not bundled or replaced. A passing cross build cannot establish their presence on every firmware (`scripts/build-device`, `scripts/package-device`).
- Locked Go modules and pinned GitHub actions/actionlint are retained. Local hooks share lint/race checks with CI; the full check includes packager policy tests. Markdown hygiene now includes the codemap (`scripts/lint`, `scripts/check`, `scripts/check-docs.py`, `.pre-commit-config.yaml`).

## Missing Critical Features

- Persistent library caching and Surprise Mode remain planned, absent features, not newly confirmed runtime regressions. They are excluded from this safety/maintenance audit and from full-MVP completion claims. No new automatic launch mode, firmware support or UI redesign was introduced (`docs/prd.md`, `docs/implementation-plan.md`, `README.md`).

## Test Coverage Gaps

**External blockers remain open, without duplicate issues:**
- [#6](https://github.com/jellydn/knulli-random-game-picker/issues/6): physical TSP buttons in both layouts, audible sustained gameplay, normal hotkey return, external pad and usable ES restoration.
- [#7](https://github.com/jellydn/knulli-random-game-picker/issues/7): effective real emulator/core override and second-system handoff against direct ES controls.
- [#8](https://github.com/jellydn/knulli-random-game-picker/issues/8): real MagicX Zero 28 dependency/display/input/audio/gameplay/return acceptance.
- [#9](https://github.com/jellydn/knulli-random-game-picker/issues/9): precise device cold/warm start, input-to-visible latency, CPU/RSS trends and long sessions. Orb benchmarks and 100-reroll fixtures do not close it.
- The public stable package has only the limited TSP evidence described in its release notes. These source changes have host/cross-build evidence only. No hardware, release, deployment or updated public-package claim follows from this audit.

---

*Concerns audit: 2026-10-25; official main and open issues #6–#9 checked before edits.*
