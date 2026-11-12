# Random Game Picker 0.4.0

Adds a separate **Play Random Game** Ports shortcut. **Random Game Picker**
remains the settings/preview app, with its existing controls and YOLO behavior.

Open the normal app, choose system/favorites and exit to save. The shortcut
uses `state/settings.json`, picks once and launches without opening SDL, artwork
or controller input. It works with YOLO on or off, without changing that value.
Recent exclusion is session-only and does not carry between shortcut runs.

Missing/invalid/incomplete settings, an unavailable saved system, empty pools,
hidden/missing/ambiguous games, discovery failures, stale selections, lost
favorite status and failed saves block submission. There is no fallback to
All Systems. The exact system/path is freshly checked, then one raw-path ES
launch request is submitted. No retry, emulator command or gamelist write is
added. Inspect `state/last-run.log` after an error; acceptance is not gameplay.

Update the complete archive, including both executable shell scripts and the
matching binary, and preserve all of `state/`. Refresh ES after manual install;
gamelist-only views need both menu paths. Store v2026.11.12 adds ownership of
both entries; update the Store before updating this package through it.
Never install the new shortcut over a 0.3.0 binary. Older assets stay unchanged.

Host gates exercise 44 real-binary no-display cases, loopback transport safety,
race tests and package modes. The SDL suite at 1280×720 and 640×480 checks UI
save → direct launch → normal app, byte-identical saved settings with YOLO on
and off, both controller layouts, 100 rerolls per size, and safety/recovery
regressions. The release workflow builds in Debian 12 and gates AArch64 ELF,
allowed shared libraries and GLIBC ≤2.34 before publishing the tested bytes.

Requires KNULLI AArch64, glibc 2.34+ and firmware SDL2/image/ttf/gfx. No new
physical-device, audio, gameplay/hotkey-return, MagicX or performance result is
claimed. Desktop fixtures and ABI checks are not handheld testing.
