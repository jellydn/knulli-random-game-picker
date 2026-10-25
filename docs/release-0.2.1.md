# Random Game Picker 0.2.1

Compatible maintenance release for the KNULLI Go/gabagool picker. The centered
artwork UI, default/Xbox layouts, saved settings schema and exact-path ES launch
contract are unchanged. No prior release, asset or tag is replaced.

## Fixes and performance

- Invalid settings now fall back as a whole instead of activating partial
  filters after a JSON type error. Actual reads are bounded to 32 KiB. A failed
  settings save still prevents launch.
- The selector redraws on events/results, not on every idle poll, and computes
  its eligible count once per redraw. Artwork remains bounded and synchronous
  on the SDL thread; the released layout and controls are retained.
- Discovery reuses path identities within each scan only and checks cancellation
  while processing games. Play still performs full fresh discovery, including
  ambiguity checks across unsupported real systems and exact system/path match.
- Package ABI failures cannot be disabled by Python optimization. The highest
  required GLIBC symbol must remain at or below 2.34. `build-info.txt` records
  exact toolchain, Debian package, target SDL and Go module build inputs.
- Expanded HTTP, configuration, symlink, XML/settings boundary, write-failure
  and artwork regressions. Both controller layouts run at both target sizes.

Synthetic host discovery benchmarks at 10k/25k/50k games showed median scan
times of 166→113 ms, 420→304 ms and 988→709 ms; 50k allocations fell from about
246 MB to 188 MB per scan. These orb measurements exclude real ES, SD cards and
artwork; they are not device latency, RSS or performance-budget sign-off.

## Install or update

Verify `random-game-picker-0.2.1-knulli-aarch64.zip` against `SHA256SUMS.txt`.
Extract the complete flat archive into `RandomGamePicker` inside your configured
Ports root. Keep the launcher, binary, assets and licenses together; both
`RandomGamePicker.sh` and `random-game-picker` must be executable. Preserve the
entire `state/` directory. Settings and menu path are unchanged from 0.2.0.

Requires KNULLI Linux AArch64, glibc **2.34+**, firmware SDL2, SDL2_image,
SDL2_ttf and SDL2_gfx, and the local ES HTTP API during the Port session.
Firmware libraries are not bundled or replaced. `artifact.json`, `abi.txt`,
`archive-inventory.txt`, `SHA256SUMS.txt` and `build-info.txt` describe the exact
release package and its build inputs. Never reuse another version's digest/sizes.

## Verification and remaining limits

Host checks include format/modules/vet, race tests, HTTP contracts, package ABI
policy regressions, shell/workflow/document lint and hooks. Normal and
race-enabled SDL fixtures exercise both layouts at 1280×720 and 640×480, 100
rerolls per size, saved/empty filters, delayed/failed discovery and recovery,
malformed/oversized art, stale-selection blocking and SDL teardown before one
exact-path POST. Race-enabled SDL fixtures are a separate local verification.
The gated workflow repeats host checks (including Go race tests), normal SDL
fixtures and ARM64 package checks before publishing the tested archive without
rebuilding it.

No new hardware validation is claimed. The limited TSP development-build
evidence belongs to [0.2.0](https://github.com/jellydn/knulli-random-game-picker/releases/tag/v0.2.0),
not a physical-device run of 0.2.1. Physical controls, audible sustained play,
normal hotkey return, external pads, real overrides/second-system handoff,
MagicX and precise device measurements remain tracked in issues
[#6](https://github.com/jellydn/knulli-random-game-picker/issues/6),
[#7](https://github.com/jellydn/knulli-random-game-picker/issues/7),
[#8](https://github.com/jellydn/knulli-random-game-picker/issues/8) and
[#9](https://github.com/jellydn/knulli-random-game-picker/issues/9).

There is no persistent library cache, async artwork pipeline or Surprise Mode.
HTTP acceptance means submitted, not played. Inspect logs before manually
retrying an unknown launch outcome; there is no automatic retry or emulator
command fallback. Rollback uses a complete prior archive while preserving state.
