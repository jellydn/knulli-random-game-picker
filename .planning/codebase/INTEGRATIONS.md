# External Integrations

**Analysis Date:** 2026-10-25

## APIs & External Services

**Local EmulationStation API:**
- KNULLI EmulationStation (ES) - Supplies the live game catalogue and receives one game launch request. The application has no Internet service dependency. Evidence: `internal/library/library.go`, `internal/launcher/es.go`.
- SDK/Client: Custom Go transport using `net`, `net/http`, and `net/url`. It opens a fresh TCP4 connection to `127.0.0.1:1234` for each request. There is no DNS lookup, proxy, redirect, or automatic retry. Evidence: `internal/launcher/es.go`.
- Auth: None. The endpoint is fixed to loopback; no token or public Web Access setting is required. Evidence: `internal/launcher/es.go`, `docs/testing.md`.
- Requests: `GET /systems`, `GET /systems/{escaped-system-id}/games`, and `POST /launch`. Launch sends the exact ES game path as `text/plain`, not an emulator command. Evidence: `internal/library/library.go`, `internal/launcher/es.go`.
- Contract: Five-second dial timeout and connection deadline; 2xx status; known Content-Length up to 64 MiB; no transfer encoding. Incomplete responses fail. An error after submission can leave the launch outcome unknown. Evidence: `internal/launcher/es.go`, `tests/http-contract.py`.
- Launch sequence: Read the catalogue again, verify the selected system/path, close SDL and controllers, save filters, then submit once. HTTP acceptance does not prove gameplay. Evidence: `internal/ui/run.go`, `internal/picker/state.go`, `cmd/random-game-picker/main.go`.

**Firmware and controller services:**
- KNULLI Ports/configgen - Runs the app from Ports and must handle the queued ES game after the Port exits. The picker does not start emulator binaries or nest configgen. Evidence: `packaging/RandomGamePicker.sh`, `internal/launcher/es.go`, `docs/testing.md`.
- SDK/Client: SDL GameController through go-sdl2 and gabagool. Profile policy loads `/tmp/gamecontrollerdb.txt`. Evidence: `internal/firmware/profile.go`, `internal/ui/run.go`, `internal/input/mapping.go`.
- Auth: None. The wrapper reads `/var/run/ports_xbox_layout.flag` and selects the KNULLI or Xbox face-button policy. It does not edit system mappings. Evidence: `packaging/RandomGamePicker.sh`, `internal/input/mapping.go`.
- Firmware scope: Only `knulli` is accepted. There is no second firmware adapter in the checkout. Evidence: `internal/firmware/profile.go`.

**Build-time distribution services:**
- Go download site and Go module distribution - Supply checksum-verified Go 1.26.0 and versioned modules. SDK/Client: `curl` and `go mod download`. No application credential is configured. Evidence: `.agents/setup`, `go.mod`, `go.sum`.
- Debian/Ubuntu repositories - Supply host and cross-build packages through APT. These are build-time services, not runtime APIs. Evidence: `.agents/setup`, `scripts/build-device`.
- GitHub Releases - Hosts downloadable packages and audit files. SDK/Client: GitHub CLI in the publication job. Auth: `GH_TOKEN` from `github.token`. Evidence: `.github/workflows/release.yml`.

## Data Storage

**Databases:**
- None. Library and selection data use Go structs and slices. Persistent filters use a local JSON file. Evidence: `go.mod`, `internal/library/library.go`, `internal/picker/state.go`.
- Connection: No database environment variable or connection string. Evidence: `internal/firmware/profile.go`, `internal/picker/state.go`.
- Client: Go standard `encoding/json`, `encoding/xml`, and `os`; no ORM. Evidence: `internal/library/library.go`, `internal/picker/state.go`.

**File Storage:**
- Local filesystem only. The picker reads ROM existence, ES XML configuration, gamelists, and existing artwork; it does not write these inputs. Evidence: `internal/library/library.go`, `internal/ui/canvas.go`.
- Base configuration priority: user `es_systems_custom.cfg`, user `es_systems.cfg`, shared `es_systems.cfg`, then legacy `es_systems.cfg`. User/shared `es_systems_*.cfg` files add or override systems. Evidence: `internal/library/library.go`.
- Default directories: user `/userdata/system/configs/emulationstation`, shared `/usr/share/emulationstation`, and legacy `/etc/emulationstation`. `--es-config` overrides the user directory. `HOME` supplies home-relative expansion, with `/userdata/system` as fallback. Evidence: `internal/firmware/profile.go`, `cmd/random-game-picker/main.go`.
- Gamelists: `<configured-system-root>/gamelist.xml`, then `<user-config>/gamelists/<system>/gamelist.xml`. Each XML read is limited to 32 MiB. Live ES data supplies launch paths and favorite flags. Evidence: `internal/library/library.go`.
- Artwork: Local image, thumbnail, or marquee path from the gamelist. UI accepts PNG/JPEG up to 8 MiB and 2048×2048 pixels. No artwork download service is used. Evidence: `internal/library/library.go`, `internal/ui/canvas.go`.
- State: `settings.json` stores only `system` and `favorites`. Save writes `settings.json.tmp` with mode 0600, then renames it. New state directories use mode 0700. Evidence: `internal/picker/state.go`.
- State location: The wrapper sets `<package>/state`; direct execution defaults to `/userdata/system/configs/random-game-picker`. `PICKER_STATE_DIR` overrides the default. Preserve state on package update. Evidence: `packaging/RandomGamePicker.sh`, `internal/firmware/profile.go`, `docs/testing.md`.
- Package data: Bundled `assets/DejaVuSans.ttf`, module notices, and license files. Firmware shared libraries are not included. Evidence: `scripts/package-device`, `THIRD_PARTY.md`.

**Caching:**
- No persistent library cache or external cache service. Startup, manual rescan, and launch verification read the catalogue. Selection uses an in-memory library and pool. Evidence: `internal/ui/run.go`, `internal/library/library.go`, `docs/testing.md`.
- Path identities are reused only within one discovery call. A new Play/rescan resolves them again. Selector idle frames reuse the displayed result until an SDL event or state result arrives. Evidence: `internal/library/library.go`, `internal/ui/run.go`.
- Recent selections retain at most ten paths in session memory. Artwork retains one selected texture until reset. Neither is a disk cache. Evidence: `internal/picker/state.go`, `internal/ui/canvas.go`.

## Authentication & Identity

**Auth Provider:**
- None. There are no user accounts, login screens, OAuth clients, or application API keys. Evidence: `go.mod`, `cmd/random-game-picker/main.go`, `internal/launcher/es.go`.
- Implementation: The app uses local firmware paths and a fixed loopback endpoint. Game identity uses exact ES paths and canonical paths to exclude ambiguous entries, not a user identity service. Evidence: `internal/library/library.go`, `internal/launcher/es.go`.
- CI publication credentials are separate from runtime identity. Evidence: `.github/workflows/release.yml`.

## Monitoring & Observability

**Error Tracking:**
- No remote error tracker or telemetry integration is configured. Errors remain local. Evidence: `go.mod`, `cmd/random-game-picker/main.go`, `packaging/RandomGamePicker.sh`.

**Logs:**
- The Ports wrapper replaces `state/last-run.log` each session and redirects stdout/stderr there. The app reports firmware/button policy and launch submission status. Evidence: `packaging/RandomGamePicker.sh`, `cmd/random-game-picker/main.go`.
- gabagool writes `ui.log` in the selected state directory. Demo mode uses the OS temporary file `picker-demo-ui.log`. Evidence: `internal/ui/run.go`.
- CI captures UI evidence under `build/ui-smoke/` and uploads it as an artifact. Package checks write ELF, digest, size, and inventory reports under `build/release/`. Evidence: `.github/workflows/check.yml`, `scripts/package-device`.

## CI/CD & Deployment

**Hosting:**
- Native app on KNULLI Linux AArch64; no hosted backend. Manual deployment extracts the complete flat archive to `RandomGamePicker` within the configured Ports root. Evidence: `scripts/package-device`, `packaging/RandomGamePicker.sh`, `docs/testing.md`.
- The standard experimental App Store path is `/userdata/roms/ports/RandomGamePicker`. This checkout documents that integration but contains no Store manifest or Store client. Custom Ports roots need manual installation. Evidence: `docs/testing.md`, `README.md`.
- GitHub published `v0.2.0` on 2026-10-24 with `random-game-picker-0.2.0-knulli-aarch64.zip`, `SHA256SUMS.txt`, `artifact.json`, `abi.txt`, and `archive-inventory.txt`. Evidence: [release record](https://github.com/jellydn/knulli-random-game-picker/releases/tag/v0.2.0).
- Published `v0.2.0` requires glibc 2.34+ and SDL2/image/ttf/gfx. The checkout now matches official main and rejects required GLIBC symbols above 2.34, including when Python optimization is enabled. Build input versions are recorded in `build-info.txt`. Evidence: `scripts/package-device`, `scripts/build-device`, `tests/package-device.py`.

**CI Pipeline:**
- GitHub Actions check workflow: Pull requests, pushes to `main`, and manual dispatch run host checks/UI tests plus a Debian Bookworm AArch64 package build. Permissions are read-only; obsolete runs are cancelled. Evidence: `.github/workflows/check.yml`.
- Host checks: Shell syntax, ShellCheck, local document/workflow checks, Go format/module verification/vet/race tests, host build, test shim build, and a real loopback HTTP fixture. Evidence: `scripts/check`, `scripts/check-docs.py`.
- Checkout release workflow: Trigger is only tag `v0.2.0`. Host/device jobs must pass; publication verifies tag/version/asset identity and checksums, then uploads those exact bytes without rebuilding. Only publication has `contents: write`. Evidence: `.github/workflows/release.yml`.
- Version boundary: Source and packager use `0.2.0`. Local audit changes do not replace the stable public archive or authorize publication. Physical acceptance and precise device measurements remain in issues #6–#9. Evidence: `cmd/random-game-picker/main.go`, `README.md`, `.planning/codebase/CONCERNS.md`.

## Environment Configuration

**Required env vars:**
- No user-provided application secret or mandatory environment variable. The Ports wrapper sets `PICKER_STATE_DIR` and `PICKER_CONTROLLER_LAYOUT` for installed use. Evidence: `packaging/RandomGamePicker.sh`, `internal/firmware/profile.go`.
- `HOME` controls ES home-relative paths. Optional `PICKER_STATE_DIR` controls app-owned settings and logs. Optional `PICKER_CONTROLLER_LAYOUT=knulli` enables the default KNULLI translation; the wrapper selects `xbox` when the firmware flag exists. Evidence: `internal/firmware/profile.go`, `packaging/RandomGamePicker.sh`.
- UI-owned variables: `FALLBACK_FONT`, `ENVIRONMENT`, `WINDOW_WIDTH`, and `WINDOW_HEIGHT`. UI setup clears `FLIP_FACE_BUTTONS`. Evidence: `internal/ui/run.go`.
- Build variables: `GOFLAGS`, `CGO_ENABLED`, `GOOS`, `GOARCH`, `CC`, and `PKG_CONFIG_LIBDIR` are set by scripts. Optional test variable `PICKER_UI_RECORD` enables an ffmpeg recording. Evidence: `scripts/check`, `scripts/build-device`, `tests/ui-smoke.py`.
- Publication variables: `GH_TOKEN`, `GH_REPO`, and `TAG` come from GitHub Actions context. Evidence: `.github/workflows/release.yml`.

**Secrets location:**
- No runtime secrets file is configured. GitHub supplies the publication job token as `github.token`; no custom release secret is named. Checkout steps disable persisted credentials. Evidence: `.github/workflows/check.yml`, `.github/workflows/release.yml`, `docs/development.md`.

## Webhooks & Callbacks

**Incoming:**
- No application HTTP server or incoming webhook. Fake HTTP servers exist only in tests. GitHub repository events trigger CI, not a device endpoint. Evidence: `cmd/random-game-picker/main.go`, `tests/http-contract.py`, `tests/ui-smoke.py`, `.github/workflows/check.yml`.

**Outgoing:**
- No Internet webhook or completion callback. The only runtime HTTP destination is the local ES API, including `POST /launch`. The app exits after submission and receives no gameplay completion event. Evidence: `internal/launcher/es.go`, `cmd/random-game-picker/main.go`.

---

*Integration audit: 2026-10-25*
