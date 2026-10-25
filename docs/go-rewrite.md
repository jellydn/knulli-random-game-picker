# Go rewrite — KNULLI first

## Research and selected patterns

Studied [Grout](https://github.com/rommapp/grout/tree/15b02e269980d05768807686ad9ecbbd0879229c)
and gabagool/v2 `v2.26.6`, as well as the picker's existing discovery, HTTP,
controller and UI contracts. Relevant Grout sources are `cfw/firmware.go`,
`cfw/cfw.go`, `app/setup.go`, `app/router.go`, `ui/game_details.go`,
`ui/setting_row.go` and `ui/footer_helpers.go`.

Adapted: explicit startup profile selection, firmware-owned paths/input rules,
typed view transitions, focus retention, contextual hints, gabagool's console
settings list and resource lifetime. The owner chose concept B: centered artwork,
game title, system and eligibility count, with controls at the bottom.

Not adapted: RomM accounts, downloads, sync, save naming, platform folder guesses,
automatic update or Grout's entire router/firmware registry. These do not serve
random selection from an existing KNULLI library. No Grout implementation or
artwork is copied. Gabagool is a pinned dependency, not vendored source.

## Boundaries

| Package | Owns | Does not own |
|---|---|---|
| `internal/library` | ES configuration/XML, live catalogue reconciliation, metadata, eligibility, uniform choice | SDL, emulator commands, writes |
| `internal/launcher` | Five-second, bounded loopback request, exact plain-text path submission | Retry, proxy, redirect, external URLs |
| `internal/firmware` | KNULLI paths, state location, controller database, launch capability | UI or scanning implementation |
| `internal/input` | SDL logical button → semantic action and gabagool virtual map | Raw device button tables |
| `internal/picker` | Selector/settings/exit, operation state, filters, recent choices, focus, revalidation | SDL or HTTP |
| `internal/ui` | Gabagool lifecycle/settings and compact selector rendering | Launch submission |
| `cmd/random-game-picker` | Startup, backend wiring, saved state, handoff after UI return | Emulator command construction |

The custom selector uses gabagool's SDL window because its stock detail screen
has only two configurable action buttons, while the picker needs A, X, Y, B and
Start. Settings use the stock gabagool list. Left/right map to its cycle callbacks
only while settings is open. A selects/toggles as before. The pending scan applies
on the UI thread and cannot navigate away from settings.

## Safety and compatibility

- Startup selects, never launches. Filters use the same `system`/`favorites` JSON
  keys as 0.1.1. Recent exclusion remains optional and session-only, with no fallback.
- Configured ES roots and gamelist artwork are retained. Live ES favorites win.
  Ports, collections, hidden/missing games and duplicate launch paths are excluded.
- Play rechecks the live catalogue and matches both system and exact path. All
  UI, fonts, images and input resources close before one `POST /launch`.
- Timeout or incomplete response is an unknown launch outcome, not retry permission.
  No shell, emulator command, configgen nesting, metadata writes or external network.
- Keep the existing launcher and its `/var/run/ports_xbox_layout.flag` policy.
  Explicit mappings avoid gabagool's additional default face-button swap.
- This is KNULLI-only. Unknown profiles fail before opening the UI. A future
  profile needs its own discovery/revalidation/launch contract, controller and
  display policy, package/ABI check and real device proof. `CanLaunch=false`
  disables the selector action; a filesystem layout alone never enables it.

## Delivery and rollout

The source identifies itself as `0.2.0`; it does not replace `0.1.1` assets.
The owner accepted the TSP development-build evidence and approved this release.
The `v0.2.0` tag runs host/device gates before publishing the exact built package.
Go/cgo adds a firmware-provided SDL2_gfx dependency. No firmware libraries are
bundled or overwritten. See [release notes](release-0.2.0.md) for evidence and limits.

The owner must retest on each handheld: physical buttons, boot/quit, real launch
and return, configured emulator overrides, saves/history, and cold-start performance.
Rollback means reinstalling the complete published 0.1.1 archive, preserving `state`.

## App Store asset contract

The Store entry is `catalogue/packages/io.github.jellydn.random-game-picker.json`
in knulli-app-store. Update its published `0.1.1` contract only after verifying
the public `v0.2.0` asset against that release's checksums and metadata. Never pin
a local development ZIP or use a mutable `latest` download alias.

The current package contract is:

| Field | Go 0.2.0 release |
|---|---|
| Version / asset | `0.2.0` / `random-game-picker-0.2.0-knulli-aarch64.zip` |
| Pinned URL | `https://github.com/jellydn/knulli-random-game-picker/releases/download/v0.2.0/random-game-picker-0.2.0-knulli-aarch64.zip` |
| Archive | Flat ZIP, `strip_components=0`; no enclosing directory |
| Files | `RandomGamePicker.sh`, `random-game-picker`, `assets/DejaVuSans.ttf`, `licenses/`, `LICENSE`, `README.md`, `THIRD_PARTY.md` |
| Destination | `/userdata/roms/ports/RandomGamePicker` (standard Ports root only) |
| Executables | `RandomGamePicker.sh`, `random-game-picker` |
| Menu path | `./RandomGamePicker/RandomGamePicker.sh` (unchanged) |
| Preserve | Entire `state/` directory on repair, update, uninstall and rollback |
| State schema | `settings.json`: `system` string and `favorites` boolean, unchanged from 0.1.1; atomic `settings.json.tmp` |
| Logs | Per-session `state/last-run.log`, plus new toolkit `state/ui.log` |
| Runtime | Linux AArch64; glibc 2.34+; `/lib/ld-linux-aarch64.so.1`; system SDL2/image/ttf/**gfx** and libc |
| Metadata | `artifact.json`, `SHA256SUMS.txt`, `archive-inventory.txt`, `abi.txt` from the exact packaged bytes |

For historical device-test traceability, the development ZIP had SHA-256
`2411cbae57f44897ef1e69cc212a480f7fd8f8c5c64fbf235ecb6f6ac72c2c6c`,
3,278,165 compressed bytes and 7,878,650 installed regular-file bytes. Its highest
required GLIBC symbol is **2.34**; SDL2_gfx is a dynamic dependency. These values
describe that development artifact, not the stable release. The stable ZIP has
fresh metadata from its exact bytes; do not reuse the development digest or sizes.

The Store update must change version, pinned release/checksum/ABI URLs,
SHA-256, both sizes, exact ELF-derived minimum glibc, dependency list (`sdl2-gfx`),
license evidence (Go/module notices replace C++ runtime/header notices), warnings
and device-test evidence together. Retain the existing destination, menu path,
executables, network declaration and preserve rule. Store dependency detection
must recognize firmware SDL2_gfx before offering installation. Use the published
`artifact.json` and `SHA256SUMS.txt` under the same `v0.2.0` release URL prefix.
