# KNULLI Random Game Picker

Pick a random game from your existing KNULLI library. Filter by system or
favorites, view scraped artwork, and reroll until you find a game to play.
The app asks EmulationStation (ES) to launch the game. It does not build emulator
commands or change ROMs, gamelists, artwork, or saves.

## Platform and release

[Version 0.2.0](https://github.com/jellydn/knulli-random-game-picker/releases/tag/v0.2.0)
is the published Go/gabagool release. It requires:

- KNULLI on Linux AArch64, with glibc **2.34 or later**.
- Firmware libraries: SDL2, SDL2_image, SDL2_ttf, and **SDL2_gfx**.
- The local ES HTTP API at `127.0.0.1:1234`, available while the Port runs.

Firmware libraries are not included or replaced. No other firmware is supported.
The target devices are TrimUI Smart Pro (1280×720) and MagicX Zero 28 (640×480).
The development build has limited TSP test evidence; MagicX remains unverified.
See [device evidence and test limits](docs/testing.md).

This checkout still identifies itself as `0.2.0-dev`. Its package script uses
that version, and its release workflow still targets `v0.1.1`. A local build is
not the published `v0.2.0` asset.

## Install and use

1. Download `random-game-picker-0.2.0-knulli-aarch64.zip` and `SHA256SUMS.txt`
   from the [v0.2.0 release](https://github.com/jellydn/knulli-random-game-picker/releases/tag/v0.2.0).
   Verify the archive, for example with `sha256sum -c SHA256SUMS.txt`.
2. Create `RandomGamePicker` inside your **configured Ports root**. Extract the
   complete archive there. Keep `RandomGamePicker.sh`, `random-game-picker`,
   `assets/`, and the license files together. Do not put ROMs in this directory.
3. Make the launcher and binary executable:
   `chmod +x RandomGamePicker.sh random-game-picker`.
4. Refresh the ES game list if necessary. Open **Random Game Picker** from Ports.
   After discovery, the app shows a game. Startup never launches a game.

For an update, replace the complete package but preserve the entire `state/`
directory. The settings schema is unchanged from 0.1.1. The release also provides
`artifact.json`, `abi.txt`, and `archive-inventory.txt` for package inspection.

## Controls

| Input | Game selector | Settings |
|---|---|---|
| A | Play the selected game | Select System, or toggle/reset the selected setting |
| X | Pick another game | — |
| Y | Open settings | Rescan the library |
| B | Return to settings | Quit |
| Start | Quit | Quit |
| D-pad | — | Up/down: choose a row; left/right: change its value |

Opening settings does not rescan. Play checks the live catalogue again, closes
the UI, then submits the exact game path once. HTTP acceptance means
**submitted**, not played. If a launch times out or returns unexpectedly, inspect
the log before you try again.

Button names follow ES inputs. The Port launcher applies the KNULLI/Xbox-layout
policy. Custom mappings and external controllers need separate tests. See the
[controller and launch guide](docs/testing.md#controller-mapping-and-launch-safety).

## Configuration and state

Use Settings to select a system, enable favorites, exclude recent selections,
or clear recent selections. Only the system and favorites filters are saved.
Recent exclusion uses up to ten selections from the current session. If no game
is eligible, change filters or clear recent selections; the app does not silently
relax your filters.

When started through the Port launcher, app-owned files are under `state/` beside
the binary: `settings.json`, its temporary replacement, `last-run.log`, and
`ui.log`. The session log is replaced on each start. Manual updates must preserve
this directory. Without the launcher, the default state directory is
`/userdata/system/configs/random-game-picker`; `PICKER_STATE_DIR` overrides it.

## Build and test

On a Debian/Ubuntu development host:

```sh
./.agents/setup
./scripts/check
```

Setup installs the host dependencies and Go 1.26.0. The check command runs
document and shell checks, Go format/module checks, vet, race tests, a host build,
and loopback HTTP tests. The binary is `build/host/random-game-picker`.

For a desktop demo with no library scan, launch request, or saved-filter writes:

```sh
./build/host/random-game-picker --demo --windowed --size 640x480
```

Use arrows, Enter (A), Escape (B), X, Y, and F10 (quit). See
[development and CI](docs/development.md) for SDL smoke tests and the Debian 12
AArch64 cross-build. Host tests do not prove handheld compatibility.

## Limits and further guidance

- Only games in the running ES catalogue are eligible. Ports, virtual
  collections, hidden/missing games, and ambiguous launch paths are excluded.
- No persistent library cache, Surprise Mode, downloads, or self-updater.
- Existing PNG/JPEG artwork is used when valid; otherwise a fallback is shown.
- Physical buttons, normal emulator return, MagicX, external controllers, and
  full performance targets still need device tests. See the
  [device test checklist](docs/testing.md#report-each-device-separately).

For contributors: [architecture](docs/go-rewrite.md),
[current seven-document codemap](.planning/codebase/ARCHITECTURE.md),
[implementation plan](docs/implementation-plan.md), and [PRD](docs/prd.md).
The plan and PRD describe goals, not completed feature claims.

MIT licensed; see [LICENSE](LICENSE) and [third-party notices](THIRD_PARTY.md).
No ROMs, commercial artwork, accounts, or telemetry are included.
