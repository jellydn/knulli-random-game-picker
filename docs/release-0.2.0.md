# Random Game Picker 0.2.0

Go/gabagool rewrite for KNULLI, with centered artwork, a selector-first flow and
contextual controller help. Retains saved system/favorites filters, random
selection, reroll, session recent exclusion and one exact-path ES launch request
after SDL closes. No other firmware support is claimed.

## Install or update

Download `random-game-picker-0.2.0-knulli-aarch64.zip` and verify it against
`SHA256SUMS.txt`. Extract the complete flat archive into `RandomGamePicker` inside
your configured Ports root. Keep `RandomGamePicker.sh`, `random-game-picker`,
`assets/` and notices together; make both executables executable. Preserve the
entire existing `state/` directory. The settings schema is unchanged from 0.1.1;
the new toolkit log is `state/ui.log`.

Requires KNULLI, Linux AArch64, glibc **2.34+**, and firmware SDL2, SDL2_image,
SDL2_ttf and **SDL2_gfx**. Firmware libraries are not bundled or replaced. The
local ES HTTP API at `127.0.0.1:1234` must be available during the Port session.
`artifact.json` records exact sizes, digest and ELF dependencies; `abi.txt` and
`archive-inventory.txt` provide the package audit data.

## Verification and limits

Host race/HTTP/SDL fixtures cover both layouts and target resolutions, 100
rerolls per size, saved filters, empty/failed/delayed discovery, stale-selection
blocking and SDL teardown before a single exact-path request.

The same application implementation was tested as a development build on TSP,
KNULLI beta 2026/09/21, 1280×720: SDL2_gfx resolved, selector/settings rendered,
injected events through the real evdev controller path exercised default and
effective Xbox mappings, 100 paced rerolls completed at about 36 MiB RSS, and an
exact-path Game Boy handoff initialized Gambatte, audio and input. The emulator
later exited with RetroAchievements login errors; ES display was restored.
The stable build changes the version string and package documentation, not the
picker implementation.

This is not proof of physical button presses, audible sound, uninterrupted
gameplay or normal hotkey return. MagicX, external controllers, emulator
overrides, a second emulated system and precise cold-start/latency/FPS budgets
remain unverified. No persistent library cache or Surprise Mode is included.
HTTP acceptance means submitted, not played. Inspect logs before retrying a
launch with an unknown outcome.

The 0.1.1 release and its assets are unchanged. Rollback uses its complete
archive while preserving `state/`.
