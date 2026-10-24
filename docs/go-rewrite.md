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

The Go rewrite is published as
[v0.2.0](https://github.com/jellydn/knulli-random-game-picker/releases/tag/v0.2.0).
This checkout still identifies itself as `0.2.0-dev`, and its release trigger
targets only the historical `v0.1.1` tag. Local builds must not be presented as
the published asset. A future release needs a separate reviewed version/trigger
change. Go/cgo adds a firmware-provided SDL2_gfx dependency. No firmware libraries
are bundled or overwritten. Earlier release assets are unchanged.

See the [device guide](testing.md) for the limited TSP development-build evidence
and the remaining tests: physical buttons, normal launch/return, emulator
overrides, saves/history, and cold-start performance on each handheld. MagicX
remains unverified. Rollback means reinstalling the complete published 0.1.1
archive, preserving `state`.
