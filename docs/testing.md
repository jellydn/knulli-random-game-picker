# Installation and testing — 0.2.0

## Go/gabagool release

The current source rewrites the app in Go with gabagool/v2 and the selected
centered-artwork layout. It retains the controls, state keys and KNULLI launch
contract below. It now requires firmware-provided **SDL2_gfx** in addition to
SDL2/image/ttf, and writes a separate `state/ui.log` for toolkit diagnostics.
Do not install on a device without checking the new ELF dependencies. TSP
development-build testing covers dependency resolution, rendering, injected
controller events, rerolls, exact-path handoff and ES restoration. Physical
buttons, audible audio, normal gameplay/hotkey return and MagicX remain unverified.
Only KNULLI is supported; `--firmware` rejects other names. See
[Go architecture](go-rewrite.md) and [development](development.md).

## Requirements and scope

This build targets TrimUI Smart Pro (1280×720) and MagicX Zero 28 (640×480).
It requires KNULLI, aarch64 glibc **2.34 or later**, and the firmware's SDL2,
SDL2_image, SDL2_ttf and SDL2_gfx. The package does not replace system libraries.
Older or missing dependencies must be blocked, not bypassed. See the
[release notes](release-0.2.0.md) for exact test scope and remaining limits.

Runtime reads ES system configurations, `gamelist.xml` artwork paths, and the running ES catalogue through local `127.0.0.1:1234` HTTP. The local API must be available while a Port runs. Public Web Access and Internet access are not required. No DNS, redirects, downloads, telemetry, accounts, or non-loopback endpoints are used. The App Store manifest declares network use because loopback sockets are still network API use.

Version 0.1.1 adds selector-first startup and preserves saved filters before discovery. Host fixtures and AArch64 ABI checks cover this version; the earlier TSP report does not verify this new flow. It does not close the plan's full M0/M3 device gates. ES must queue the submitted game and process it after the Port's wrapper/configgen exits. No fallback invokes emulator binaries or nests configgen.

## Install

Prefer the reviewed experimental App Store entry when available. Its first package targets the standard KNULLI Ports layout at `/userdata/roms/ports/RandomGamePicker` and owns one menu entry. It does not rewrite ES system configuration. If your configured Ports root is different, do not use that fixed-location Store package; use the manual archive below in your actual configured Ports root.

For manual installation, verify the ZIP against the release's `SHA256SUMS.txt`, create a `RandomGamePicker` directory **inside your configured Ports root**, and extract all archive contents there. Keep `RandomGamePicker.sh`, the binary, `assets/`, and licenses together. Ensure the launcher and binary are executable; refresh ES's game list if needed. Open **Random Game Picker** from Ports. Do not place ROMs in the package directory.

The launcher resolves its own directory, inherits KNULLI's SDL mappings, and writes `state/settings.json`, its temporary replacement, `state/last-run.log` and `state/ui.log` below the package directory. The session log is replaced each run. Store repair/update/uninstall preserve `state`. ROMs, gamelists, scraped art, emulator settings, and saves are read-only to the picker. ES/emulators still perform their normal play-history/save writes when launched.

## Controls

### Version `0.2.0` — selector-first flow retained from `0.1.1`

After discovery, the selector shows a random game using your saved system and favorites filters. Startup does not launch a game. Press A once to play, X to reroll, or Y to open the existing settings/filters screen. If the library or filters have no eligible games, Play is unavailable; Y still opens settings so you can change filters or rescan.

- In settings, D-pad up/down chooses a row; left/right on System changes the console.
- A on System picks a game and returns to the selector. A on other rows toggles or resets that setting.
- Y inside settings rescans; opening settings alone does not scan.
- B from the selector returns to settings; B from settings quits. Start still quits globally.
- The saved system and favorites values are applied before the first selection and retained on exit. Recent exclusion remains session-only.

Retest first-press Play, Y settings/rescan, B back/quit, and saved filters on each handheld. Install the complete `0.2.0` archive and preserve `state`; do not replace just the binary. Earlier release assets remain unchanged.

### Published `0.1.0` — filters-first flow

- D-pad up/down: choose a filter row.
- Left/right on System: cycle All Systems or one console.
- A on System: select a game. A on other rows: toggle/reset.
- Preview A: submit Play; X: reroll; B: return to filters.
- Y: rescan; B from filters or Start: quit.

### Controller mapping and launch safety

Alpha.1 reversed A/B and X/Y on the owner's TrimUI Smart Pro. Alpha.2 translates KNULLI's default generated SDL mapping back to ES button labels. The launcher reads `/var/run/ports_xbox_layout.flag`, the same mode flag used by KNULLI configgen, and leaves the non-swapped Xbox-layout mapping unchanged. Update **both the launcher and binary**, not just the binary. Store updates preserve `state`; manual updates must also leave it intact.

Buttons still use the inherited SDL GameController mapping (also loading KNULLI's `/tmp/gamecontrollerdb.txt`); no raw device button indices are hardcoded. The session log reports the selected layout. The labels describe ES inputs, so custom ES mappings or differently labelled external pads still need testing. Direct desktop execution without the Port launcher retains standard SDL labels; `PICKER_CONTROLLER_LAYOUT=knulli` selects the default KNULLI translation for fixture testing. Report physical A pick/play, B back, X reroll and Y settings/rescan separately, and record which build you tested. Do not modify system mappings blindly.

Source contract: KNULLI [controller.py](https://github.com/knulli-cfw/knulli-linux/blob/knulli-main/package/system/knulli-configgen/configgen/configgen/controller.py) maps default ES `a/b/x/y` to SDL `b/a/y/x`; its Xbox-layout mode maps names directly. The [Ports hook](https://github.com/knulli-cfw/knulli-linux/blob/knulli-main/board/fsoverlay/usr/share/knulli/configgen/scripts/ports_hooks.sh) owns the layout flag. This is a shared Ports policy, not a guessed TSP button-number table.

Play rechecks the full live catalogue, closes SDL/input, submits the exact path **once** as plain text to ES, then exits. A 2xx response means submitted, not played. After a timeout or unexpected return to Ports, inspect the log before trying again; never assume a failed response means ES did not accept the request.

## Known limits, not completed MVP claims

- The owner confirmed that the corrected alpha.2 works on TSP. No completed launch/return matrix, configured-emulator success rate, Quick Resume/history proof, or cached-startup/performance measurements yet; MagicX remains unverified.
- No persistent game-index cache. A background ES/config/XML read runs on startup/rescan and before Play. Selection uses an in-memory flat pool; artwork decodes synchronously for the selected game. The plan's M2 budgets are not claimed.
- Recent exclusion is session-only and optional; it never silently relaxes an empty pool. Reset recent selections to continue.
- Favorites come from live ES metadata. Artwork comes from gamelist PNG/JPEG files capped at 8 MiB and 2048×2048; other/oversized/missing art uses a fallback.
- No Surprise Mode in this release. No automatic system rescans, ES metadata writes, emulator selector, downloads, or self-updater.
- Ports/virtual collections are excluded from selection; ambiguous paths shared by real systems are excluded to prevent multiple ES launches. Only games already loaded by ES are eligible.
- The HTTP client expects Content-Length responses from the inspected ES API and fails closed on unsupported transfer encodings.

## Report each device separately

Record tester/date, hardware revision, exact KNULLI image/build, CPU architecture, glibc/SDL versions, resolution, package version, and controller name/GUID. Do not send ROMs or private metadata.

1. Install; open and quit; reopen. Confirm no layout clipping, correct physical buttons, and the log/settings location.
2. Verify one console filter and favorites against ES. Preview scraped art and missing-art fallback. Complete 100 rerolls. Enable recent exclusion, exhaust a small pool, then reset it.
3. Select a known working game with a per-game emulator override. Press Play once. Record whether the correct game starts once, input/audio work, and the normal hotkey returns to a usable ES menu. Compare with launching the same game directly in ES.
4. Repeat with another system and external controller if available. Check ES play history, saves, and display restoration. Stop if the device freezes or opens multiple games.
5. Repair/update and verify settings remain; uninstall and verify the app/menu are removed while `state` and all ROM/save/art files remain. Do not delete preserved state unless you want to reset it.

For launch failures, attach the relevant non-sensitive part of `RandomGamePicker/state/last-run.log`. Include a short video of the button press, Port exit, and resulting ES/game state. A desktop fixture result is not a substitute for this report.
