# Random Game Picker 0.3.0

Optional persistent YOLO mode for the KNULLI Go/gabagool picker. It is off by
default. Existing settings files keep the normal preview flow without migration.

## Pick and launch in one step

Open Settings with Y, select YOLO mode and press A or left/right to toggle it.
Toggling does not pick or launch a game. Exit normally to save the `yolo` boolean
in `state/settings.json`; a launch also saves settings before submission.

With YOLO on, A on System picks a random eligible game and immediately starts
the existing fresh catalogue check. It closes SDL/input before sending one
exact-path launch request to EmulationStation. There is no preview confirmation,
cancel, or reroll step. Later runs open Settings first so you can change filters
or turn YOLO off. With YOLO off, startup and picks keep preview, A Play, X reroll
and B back. Recent exclusion remains session-only.

## Safety and compatibility

Empty pools, discovery failures, stale games and failed settings saves still
block launch. Demo mode never scans, launches or saves. Tools, Emulators, Ports,
collections and ambiguous launch paths remain excluded. There is no launch retry,
emulator command, nested configgen, metadata write or external network request.
HTTP acceptance is not proof of gameplay.

Requires KNULLI AArch64, glibc 2.34+ and firmware SDL2/image/ttf/gfx. Install paths,
launcher, controller policies and file ownership are unchanged. Preserve the
entire `state/` directory on update, repair, uninstall or rollback.

## Verification scope

Host checks cover formatting, modules, vet, race tests, package ABI policy and
loopback HTTP contracts. The real SDL fixture suite covers 1280×720 and 640×480,
YOLO toggle/save/reopen, one-pick launch, stale-game/save-failure blocks, disabling
and reopening into the normal preview/reroll/cancel/Play flow, demo isolation,
both controller mappings and 100 rerolls per size. Screenshots and a keyboard/
pointer toggle recording were inspected. Release gates also build and inspect
the AArch64 package before publishing those exact bytes without rebuilding.

These fixtures do not prove physical controls, audio, normal gameplay/hotkey
return, MagicX compatibility or handheld performance. Older device reports remain
scoped to the versions actually tested; no new hardware pass is claimed.

Verify `random-game-picker-0.3.0-knulli-aarch64.zip` against `SHA256SUMS.txt` and
update the complete package, including launcher, binary, assets and notices.
Earlier tags and release assets remain unchanged.
