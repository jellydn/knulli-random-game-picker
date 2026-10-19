# Installation and testing — 0.1.0

## Requirements and scope

This build targets TrimUI Smart Pro (1280×720) and MagicX Zero 28 (640×480). On 2026-10-19 the owner confirmed that alpha.2 works on TSP after the face-button correction and requested this stable release. Exact firmware details and the full matrix below were not supplied; MagicX remains unverified. It requires KNULLI, aarch64 glibc **2.36 or later**, and the firmware's SDL2, SDL2_image and SDL2_ttf. The package does not replace system libraries. Older firmware must be blocked, not bypassed.

Runtime reads ES system configurations, `gamelist.xml` artwork paths, and the running ES catalogue through local `127.0.0.1:1234` HTTP. The local API must be available while a Port runs. Public Web Access and Internet access are not required. No DNS, redirects, downloads, telemetry, accounts, or non-loopback endpoints are used. The App Store manifest declares network use because loopback sockets are still network API use.

This is the first stable release of the existing feature set, following the owner's TSP retest. It does not close the plan's full M0/M3 device gates. ES must queue the submitted game and process it after the Port's wrapper/configgen exits. No fallback invokes emulator binaries or nests configgen.

## Install

Prefer the reviewed experimental App Store entry when available. Its first package targets the standard KNULLI Ports layout at `/userdata/roms/ports/RandomGamePicker` and owns one menu entry. It does not rewrite ES system configuration. If your configured Ports root is different, do not use that fixed-location Store package; use the manual archive below in your actual configured Ports root.

For manual installation, verify the ZIP against the release's `SHA256SUMS.txt`, create a `RandomGamePicker` directory **inside your configured Ports root**, and extract all archive contents there. Keep `RandomGamePicker.sh`, the binary, `assets/`, and licenses together. Ensure the launcher and binary are executable; refresh ES's game list if needed. Open **Random Game Picker** from Ports. Do not place ROMs in the package directory.

The launcher resolves its own directory, inherits KNULLI's SDL mappings, and writes only `state/settings.json`, its temporary replacement, and `state/last-run.log` below the package directory. The log is replaced each session. Store repair/update/uninstall preserve `state`. ROMs, gamelists, scraped art, emulator settings, and saves are read-only to the picker. ES/emulators still perform their normal play-history/save writes when launched.

## Controls

- D-pad up/down: choose a filter row.
- Left/right on System: cycle All Systems or one console.
- A on System: select a game. A on other rows: toggle/reset.
- Preview A: submit Play; X: reroll; B: return to filters.
- Y: rescan; B from filters or Start: quit.

Alpha.1 reversed A/B and X/Y on the owner's TrimUI Smart Pro. Alpha.2 translates KNULLI's default generated SDL mapping back to ES button labels. The launcher reads `/var/run/ports_xbox_layout.flag`, the same mode flag used by KNULLI configgen, and leaves the non-swapped Xbox-layout mapping unchanged. Update **both the launcher and binary**, not just the binary. Store updates preserve `state`; manual updates must also leave it intact.

Buttons still use the inherited SDL GameController mapping (also loading KNULLI's `/tmp/gamecontrollerdb.txt`); no raw device button indices are hardcoded. The session log reports the selected layout. The labels describe ES inputs, so custom ES mappings or differently labelled external pads still need testing. Direct desktop execution without the Port launcher retains standard SDL labels; `PICKER_CONTROLLER_LAYOUT=knulli` selects the default KNULLI translation for fixture testing. Report physical A pick/play, B back, X reroll and Y rescan separately. Do not modify system mappings blindly.

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
