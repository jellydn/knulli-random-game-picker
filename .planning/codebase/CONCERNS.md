# Codebase Concerns

**Analysis Date:** 2026-10-24

## Tech Debt

**Release information does not match the published package:**
- Issue: Observed. GitHub published v0.2.0 on 2026-10-24. Its asset is `random-game-picker-0.2.0-knulli-aarch64.zip`. This checkout still reports `0.2.0-dev`. The release workflow accepts only `v0.1.1`. The refreshed guides distinguish these source settings from the published release.
- Files: `cmd/random-game-picker/main.go`, `scripts/package-device`, `.github/workflows/release.yml`, `README.md`, `docs/testing.md`, `docs/development.md`, `docs/go-rewrite.md`.
- Impact: A build from this checkout is not the published stable package. A v0.2.0 tag does not trigger this workflow. Do not infer the published package's version or ABI from these build settings.
- Fix approach: Compare the checkout with the published release before the next release change. Update version, trigger, and guides together. Do not replace published assets or move a published tag. Current external evidence: [v0.2.0 release notes](https://github.com/jellydn/knulli-random-game-picker/releases/tag/v0.2.0), read on 2026-10-24.

**Build gate permits a higher glibc requirement than the published package:**
- Issue: Observed. Published v0.2.0 requires glibc 2.34 and SDL2/image/ttf/gfx. The package script permits GLIBC symbol versions through 2.36. The refreshed installation guide states the published 2.34 floor.
- Files: `scripts/package-device`, `scripts/build-device`, `docs/testing.md`, `docs/development.md`.
- Impact: A future local package can pass the gate but fail on firmware that supports the published 2.34 floor. This is a build-policy gap, not evidence that the published asset needs 2.36.
- Fix approach: Agree on the supported floor. Enforce that floor against each actual ELF report and package manifest. Keep the dependency list and installation guide consistent with it.

## Known Bugs

**No picker runtime defect was confirmed by this documentation inspection:**
- Symptoms: The v0.2.0 notes report that a TSP development build handed a Game Boy game to Gambatte. The emulator later exited with RetroAchievements login errors. ES display was restored. This does not establish a picker bug or a successful complete gameplay session.
- Files: `internal/launcher/es.go`, `cmd/random-game-picker/main.go`, `docs/testing.md`; device evidence is attributed to the [v0.2.0 release notes](https://github.com/jellydn/knulli-random-game-picker/releases/tag/v0.2.0).
- Trigger: The reported test used TSP, KNULLI beta 2026/09/21, and injected real evdev input. This audit did not repeat the test or establish the cause of the emulator exit.
- Workaround: No picker workaround is established. Compare the same game with a direct ES launch and inspect non-sensitive logs. Do not classify the account error as a picker defect without that evidence.

## Security Considerations

**Local ES service is a trust boundary:**
- Risk: The picker trusts the local ES catalogue and launch API. A different process on the same address can return false metadata or receive a launch path. Loopback isolation is not service authentication.
- Files: `internal/launcher/es.go`, `internal/library/library.go`, `tests/http-contract.py`.
- Current mitigation: The endpoint is fixed at IPv4 loopback port 1234. The client does not use DNS, proxies, redirects, or automatic retries. It sends the exact path as plain text, not a shell command. Responses have size and time limits.
- Recommendations: Keep this endpoint local. Treat firmware and local processes as trusted dependencies. Do not add a configurable remote endpoint without a separate security review.

**Metadata, artwork, and state files need bounded processing:**
- Risk: Local metadata and images can be malformed or large. Logs can contain private paths. Files can change between validation and use. This is a trust limitation, not a demonstrated exploit.
- Files: `internal/library/library.go`, `internal/ui/canvas.go`, `internal/picker/state.go`, `packaging/RandomGamePicker.sh`.
- Current mitigation: XML reads are capped at 32 MiB. Titles and launch paths are bounded. Artwork checks accept PNG/JPEG up to 8 MiB and 2048 by 2048 pixels. Settings writes use a temporary file and rename. The wrapper uses a private umask for logs and binary execution.
- Recommendations: Keep these limits and test hostile local fixtures. Share only non-sensitive log sections. The artwork check opens the path before SDL opens it again; do not claim protection against concurrent file replacement.

## Performance Bottlenecks

**Full discovery runs at startup, rescan, and Play verification:**
- Problem: Observed design. There is no persistent library cache. Each scan reads configuration, live ES systems and games, and available gamelists.
- Files: `internal/library/library.go`, `internal/ui/run.go`, `cmd/random-game-picker/main.go`, `docs/implementation-plan.md`.
- Cause: Systems are fetched in sequence. Path checks and symlink resolution run across the catalogue. Play repeats discovery before it returns a launch request.
- Improvement path: Measure real large libraries first. Add a disposable validated cache only if measurements require it. Preserve live eligibility checks before launch. No precise startup or launch-latency budget is verified.

**Selector rendering repeats work on the UI thread:**
- Problem: Observed work with an unmeasured speed impact. Selected artwork is checked and decoded synchronously. Frames rebuild text surfaces and textures. Frames also rebuild eligible pools to show counts and button states.
- Files: `internal/ui/canvas.go`, `internal/library/library.go`, `internal/picker/state.go`, `internal/ui/run.go`.
- Cause: The selector redraws after a 16 ms sleep. It stores only the current artwork texture, not text or pool results. Artwork uses a header check followed by SDL decoding.
- Improvement path: Measure frame time, allocation rate, and selection delay on each device. Cache unchanged text and counts if needed. Move image work off the UI thread only with safe SDL resource ownership. The sleep is not proof of a fixed frame rate.

## Fragile Areas

**ES launch queue and Port exit sequence:**
- Files: `internal/ui/run.go`, `internal/picker/state.go`, `cmd/random-game-picker/main.go`, `internal/launcher/es.go`, `tests/ui-smoke.py`, `tests/http-contract.py`.
- Why fragile: Correct operation depends on the installed ES API and queue behavior after the Port exits. A 2xx response confirms submission, not play. A timeout can occur after ES accepts a request. Saving settings occurs before submission; a save error prevents the handoff.
- Safe modification: Retain exact system/path revalidation, UI teardown before submission, and one request with no automatic retry. Test an unwritable state directory before changing save policy. Inspect logs before a manual retry with an unknown outcome.
- Test coverage: Host fixtures check window closure, exact request bytes, stale selection blocking, and no retry after a truncated response. The reported TSP handoff is useful evidence but does not prove normal hotkey return, overrides, or all emulators.

**Controller mapping spans wrapper, firmware profile, and two UI paths:**
- Files: `packaging/RandomGamePicker.sh`, `internal/firmware/profile.go`, `internal/input/mapping.go`, `internal/input/mapping_test.go`, `internal/ui/run.go`, `tests/controller-events.cpp`, `tests/ui-smoke.py`.
- Why fragile: The wrapper reads the KNULLI Xbox-layout flag. The profile sets face-button correction. The selector handles SDL events directly, but settings use gabagool virtual mappings. A second swap can reverse actions.
- Safe modification: Update and test the full package. Keep inherited SDL mappings and one semantic action map. Test default and Xbox layouts in both views before physical-device checks.
- Test coverage: Logical SDL fixtures cover both layouts. Release notes also report injected real evdev tests on TSP. Neither method proves physical button presses, external pads, or all custom ES mappings.

**Configuration merge and path identity determine eligibility:**
- Files: `internal/library/library.go`, `internal/library/library_test.go`, `docs/implementation-plan.md`.
- Why fragile: Discovery selects a base configuration and then merges additive files from user and shared directories. It resolves symlinks when possible. Paths shared by real systems are excluded because ES can match more than one game. A request error for one live system fails the scan.
- Safe modification: Preserve ambiguity checks across unsupported real systems too. Add explicit conflicting-override and symlink fixtures before changing precedence or identity rules. Do not assume the planned precedence is already proven for each firmware.
- Test coverage: Fixtures cover a custom root, live favorites, hidden/missing games, duplicate paths through Ports, and malformed artwork XML. They do not establish all cross-directory override combinations.

## Scaling Limits

**Catalogue and response bounds:**
- Current capacity: Each XML file is limited to 32 MiB. Each HTTP response is limited to 64 MiB. Each request has a five-second connection timeout and a five-second connection deadline. Games remain in memory; there is no measured maximum game count.
- Limit: Oversized XML or unsupported HTTP framing fails closed. A response needs Content-Length; chunked or unknown-length responses are rejected. Multiple requests mean five seconds is not a total scan deadline.
- Scaling path: Profile realistic libraries before changing limits. Stream or partition data if needed, but retain bounded input and cancellation. Sources: `internal/library/library.go`, `internal/launcher/es.go`, `docs/implementation-plan.md`.

**Artwork, recent selections, and measured device scope:**
- Current capacity: One artwork texture is retained. Images are limited to 8 MiB and 2048 by 2048 pixels. Recent history holds at most 10 picks for the current session. The release notes report 100 paced TSP rerolls at about 36 MiB RSS for the development build.
- Limit: The memory observation is not a large-library limit or a leak test. Recent exclusion can empty a small pool and does not silently relax it. Recent history is not persistent.
- Scaling path: Measure RSS and latency at the plan's 10k/25k/50k library sizes. Test both target devices. Keep the device observation separate from precise performance claims. Sources: `internal/ui/canvas.go`, `internal/picker/state.go`, `docs/implementation-plan.md`, and the [v0.2.0 release notes](https://github.com/jellydn/knulli-random-game-picker/releases/tag/v0.2.0).

## Dependencies at Risk

**Firmware SDL libraries and Go/cgo build inputs:**
- Risk: The binary dynamically depends on firmware SDL2, SDL2_image, SDL2_ttf, and SDL2_gfx. The build uses Debian Bookworm packages, not a pinned target firmware sysroot. Package versions can change between builds. No specific dependency vulnerability was confirmed here.
- Impact: A passing host build or ABI report does not prove library availability or display/input behavior on every KNULLI image. SDL2_gfx resolved in the reported TSP development test; MagicX remains unverified.
- Migration plan: Retain the locked Go modules and ELF inspection. Record exact toolchain and package versions for release builds. Check actual firmware libraries and both devices before raising requirements. No dependency replacement is justified by this audit alone.
- Files: `go.mod`, `go.sum`, `.agents/setup`, `scripts/build-device`, `scripts/package-device`, `internal/ui/run.go`.

## Missing Critical Features

**Persistent library cache and measured performance gates:**
- Problem: The implementation has no persistent game index, cache invalidation, or benchmark runner that proves the plan's device budgets. This is a known scope gap, not a new regression.
- Blocks: Claims of cached startup under three seconds and selection under 200 ms on reference libraries. Sources: `internal/library/library.go`, `internal/ui/run.go`, `docs/implementation-plan.md`.

**Surprise Mode and broader hardware acceptance:**
- Problem: Surprise Mode is a planned P1 feature but is not implemented. Current acceptance evidence does not complete the two-device launch/return matrix.
- Blocks: A claim that the full proposed MVP or all P1 scope is complete. Other firmware support is intentionally absent, not a defect. Sources: `docs/prd.md`, `docs/implementation-plan.md`, `internal/picker/state.go`, `internal/firmware/profile.go`.

## Test Coverage Gaps

**Physical-device play and return:**
- What's not tested: Physical presses, audible audio, uninterrupted gameplay, normal hotkey return, MagicX Zero 28, external controllers, emulator overrides, and a second emulated system remain unverified. The current release notes report a development build on TSP KNULLI beta 2026/09/21 at 1280 by 720, injected real evdev input, 100 rerolls, and a Gambatte handoff with later account errors and ES display restoration. They say the stable build changes version text and package documentation, not picker implementation.
- Files: `tests/ui-smoke.py`, `tests/controller-events.cpp`, `docs/testing.md`, `docs/implementation-plan.md`; use the [v0.2.0 release notes](https://github.com/jellydn/knulli-random-game-picker/releases/tag/v0.2.0) for this newer device evidence.
- Risk: Host fixtures cannot detect driver, audio, firmware queue, emulator, or hotkey failures. The old alpha.2 report does not prove the current release matrix.
- Priority: High.

**Boundary inputs, write failures, and transport variants:**
- What's not tested: The inspected suites do not directly cover all XML/HTTP size boundaries, slow responses, chunked responses, malformed image dimensions, concurrent file replacement, corrupt settings, or interrupted and failed settings writes. Discovery override precedence also lacks a full conflicting-file matrix.
- Files: `tests/http-contract.py`, `internal/library/library_test.go`, `internal/picker/state_test.go`, `internal/ui/canvas.go`, `internal/launcher/es.go`.
- Risk: Safety checks or failure messages can regress without a fixture that exercises the boundary. These are coverage gaps, not confirmed runtime bugs.
- Priority: Medium.

**Precise performance and release reproduction:**
- What's not tested: No inspected test proves cold-start, latency, FPS, or memory budgets across large libraries on both handhelds. The checkout does not reproduce the stable version label or release trigger as written. Host checks do not rebuild or verify the published asset on hardware.
- Files: `tests/ui-smoke.py`, `scripts/check`, `scripts/build-device`, `scripts/package-device`, `.github/workflows/release.yml`, `docs/implementation-plan.md`.
- Risk: A stable release label or host reroll result can be mistaken for complete hardware and performance acceptance.
- Priority: High.

---

*Concerns audit: 2026-10-24*
