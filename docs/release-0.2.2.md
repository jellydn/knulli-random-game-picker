# Random Game Picker 0.2.2

Application-system eligibility fix for the KNULLI Go/gabagool picker. Tools and
Emulators no longer enter random selection, even when ES marks their entries
as games or favorites. The existing Ports exclusion remains in effect.

## Eligibility and metadata evidence

The new rule matches exact system IDs, case-insensitively: `tools` and
`emulators`. The pre-existing Ports policy is unchanged. It does not infer an
application from display titles, platform `pc`, core `sh`, hardware `tool` or
file extensions. Real games with similar titles remain eligible. Excluded systems
still contribute to canonical-path ambiguity checks, so this change cannot
make a duplicate launch path appear safe. Startup, rescan and the fresh Play
check all use the same rule. Cached/in-memory application candidates are also
rejected during selection and launch revalidation. For path-only candidates,
cleaned descendants of `/userdata/roms/tools` and `/userdata/roms/emulators`
are rejected with path-boundary checks; `toolshed` and `emulators-extra` are
not matches. Saved application-system filters reset to All Systems on discovery.

The owner's correction was **Tools**, not a game named “Toon”. Inspection used
the actual firmware metadata at KNULLI commit
[`57afa662`](https://github.com/knulli-cfw/knulli-linux/tree/57afa662009641d181f57241bcc480119b131530):

- `package/emulationstation/knulli-es-system/es_systems.yml` defines `tools`
  (Tools, hardware `tool`) and `emulators` (Emulators, hardware `emulator`).
- `board/fsoverlay/usr/share/knulli/datainit/roms/emulators/gamelist.xml`
  identifies `./PPSSPP.sh` and `./ScummVM.sh`, both with genre `Emulator`.
- The RG35XX `es_systems_tools.cfg` uses the older ID `Tools`; matching is
  case-insensitive to cover this definition without matching partial names.

Read-only device inspection also confirmed these exact IDs and roots in
`/usr/share/emulationstation/es_systems.cfg`. Tools contained Grout, PlayTime,
RetSend, RAOfflineProxy and retroshelf; Emulators contained PPSSPP and ScummVM.
No device files were changed. No per-title application blacklist is added, and
no new hardware gameplay test is claimed.

## Verification and compatibility

The discovery regression fails without the fix. It checks lower/title-case IDs,
favorites, similarly named real games and duplicate paths shared with excluded
systems. The normal host, race, transport, packager and SDL smoke gates remain
required for release. These fixture checks do not prove handheld gameplay.

The UI, controller layouts, launcher, settings schema, install paths and
exact-path single-submit ES handoff are unchanged. Requires KNULLI AArch64,
glibc 2.34+ and firmware SDL2/image/ttf/gfx. MagicX, physical controls, audible
gameplay, normal hotkey return and precise device performance remain unverified.

Verify `random-game-picker-0.2.2-knulli-aarch64.zip` against `SHA256SUMS.txt`.
Update the complete package, including launcher, binary, assets and notices.
Preserve the entire `state/` directory. Earlier tags and assets stay unchanged.
