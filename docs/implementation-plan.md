# Implementation plan — KNULLI Random Game Picker

2026-10-19 · Proposed, not implemented · C++17 / SDL2

## 1. Outcome and decision summary

Build a small, controller-first application that reads the user's existing KNULLI library, selects uniformly from eligible games, displays scraped artwork, and delegates launch to KNULLI. Package it first as a standalone Port, then submit a versioned package to the community KNULLI App Store.

The [PRD baseline](prd.md) is the requirements source. This plan does not add emulator setup, ROM management, scraping, or firmware modifications. P0 is the first playable milestone. P1 favorites, recent exclusion, and Surprise Mode follow before the proposed v1.0 candidate. Optional roulette animation is deferred.

**Critical gate:** prove that EmulationStation (ES) can accept a game launch request while the picker Port runs, then launch exactly once after that Port exits. No launch backend is proven on either target device yet. Do this before investing in the full UI.

### Recommendations for the three open questions

| Question | Recommendation | Evidence and decision gate |
|---|---|---|
| Q1: Reliable standalone launch? | Prefer an ES-owned handoff through its existing loopback HTTP `POST /launch`. Do not call emulator binaries or use Swissknife as a launcher. | The route exists in inspected ES source, but Port-exit queue processing and installed-version availability require M0 hardware proof. HTTP success does not confirm launch. If it fails, investigate configgen as a bounded fallback, not an automatic retry. |
| Q2: Multiple emulator configurations? | Inherit ES/KNULLI's current per-game, folder, system, and default policy. Do not offer an emulator/core selector in v1.0. | ES-owned launch preserves configured commands and controller assignment. A direct configgen experiment must leave emulator/core overrides unset and verify equivalent results. |
| Q3: Port or direct ES integration? | Ship a Port with no ES patch. Return to ES when the selected game exits; reopening the picker restores its filters. | This matches the requested distribution and avoids coupling UI code to a firmware fork. A native ES extension is a later proposal only if a safe standalone launch contract cannot be proven. |

**Offline interpretation to confirm before M0 approval:** loopback HTTP is local IPC; no external network, DNS, listening server, or Internet is needed. The PRD also says “no network access.” If that means no socket use at all, HTTP is not compliant. Record owner acceptance of local IPC or select a proven non-socket backend. Do not silently weaken the requirement.

## 2. Evidence and unresolved facts

Research used `knulli-cfw/knulli-linux` on `knulli-main` and its pinned ES revision `10834d004ec3de0808703b46220b6a8574d53218`. Firmware and App Store branch links below can change. At M0, record exact firmware images, source revisions, executable paths, and API responses for each device.

| Finding | Source | Consequence |
|---|---|---|
| KNULLI generates emulator configuration on launch, including per-system and per-game settings. | [KNULLI overview](https://knulli.org/about-knulli/) | Delegate policy; do not reproduce RetroArch commands. |
| ES uses loopback port 1234 by default. `/launch` accepts an exact path as raw `text/plain`, searches loaded games, and queues UI work. It has no explicit not-found or emulator-completion acknowledgement. | [HTTP route and binding](https://github.com/knulli-cfw/batocera-emulationstation/blob/10834d004ec3de0808703b46220b6a8574d53218/es-app/src/services/HttpServerThread.cpp#L590-L621) | Capability probe plus loaded-catalogue reconciliation; never equate HTTP 2xx with successful play. |
| That handler does not stop at the first path match. | Same launch handler | A path present in multiple real systems could enqueue multiple launches. Detect ambiguity and block affected paths until M0 proves a safe remedy. |
| Current Swissknife supports status/control, not ROM launch. | [knulli-es-swissknife](https://github.com/knulli-cfw/knulli-linux/blob/knulli-main/package/system/knulli-scripts/scripts/knulli-es-swissknife) | Do not invent a Swissknife launch flag or kill the picker launch tree. |
| ES captures controllers, closes resources, runs the Port synchronously, then restores resources and updates game history. | [FileData launch lifecycle](https://github.com/knulli-cfw/batocera-emulationstation/blob/10834d004ec3de0808703b46220b6a8574d53218/es-app/src/FileData.cpp#L693-L812) | Waiting inside the Port for ES to launch can deadlock the handoff. |
| Shell Ports get SDL mappings, not the original ES controller argument vector. | [shGenerator](https://github.com/knulli-cfw/knulli-linux/blob/knulli-main/package/system/knulli-configgen/configgen/configgen/generators/sh/shGenerator.py) | Picker controls and selected-game controller forwarding are separate contracts. |
| Configgen exposes `emulatorlauncher -system <id> -rom <path>` and controller descriptors. | [argument parser](https://github.com/knulli-cfw/knulli-linux/blob/knulli-main/package/system/knulli-configgen/configgen/configgen/emulatorlauncher.py#L545-L581) | Possible fallback through KNULLI, but nested hooks, evmapy, display restoration, and missing ES bookkeeping must be tested. |
| Both devices have A133 build support in current source. | [boot device list](https://github.com/knulli-cfw/knulli-linux/blob/knulli-main/package/boot/uboot-a133/uboot-a133.mk), [MagicX board](https://github.com/knulli-cfw/knulli-linux/tree/knulli-main/board/allwinner/a133/magicx-zero-28) | Shared SoC does not prove ABI, stable release availability, or working controls. |
| The identified App Store is a community project, distinct from firmware `knulli-store`. | [App Store README](https://github.com/jellydn/knulli-app-store), [firmware store](https://github.com/knulli-cfw/knulli-linux/blob/knulli-main/package/system/knulli-scripts/scripts/knulli-store) | Target the community manifest contract; do not claim official firmware-store approval. |

## 3. Architecture and ownership

Use one executable, a thin Port wrapper, and one background worker. Keep domain code independent of SDL and KNULLI so it can run against fixtures. No plugin framework, database server, or local web server is needed.

```text
ES configuration + filesystem + gamelist metadata
                       |
              Discovery -> GameIndex -> Filter -> Uniform picker
                               ^                      |
                         Cache snapshot           Preview (SDL2)
                                                      |
                          ES launch adapter <- Launch request
                                   |
                      Release SDL -> submit once -> exit Port
                                   |
                    ES resumes -> configured game -> ES menu
```

| Module | Owns | Does not own |
|---|---|---|
| `discovery` | Effective system definitions, filesystem candidates, XML metadata, path rules | Emulator commands or ROM writes |
| `library` | Immutable index snapshots, eligibility, deduplication, uniform selection | SDL objects or subprocesses |
| `storage` | Versioned cache, JSON settings and selection history, atomic writes | ES metadata updates |
| `platform` | Firmware capabilities, loaded ES catalogue, launch handoff | Emulator configuration policy |
| `ui` | Screen state, SDL events, artwork textures, focus and prompts | Blocking scans or launch command assembly |
| `app` | Worker lifetime, state transitions, error presentation and shutdown | A second model of system settings |

Suggested dependencies: SDL2, SDL2_image (JPEG/PNG; WebP only if available), SDL2_ttf, pugixml, and nlohmann/json. Prefer firmware-compatible libraries and pinned build inputs. Select the smallest existing HTTP client available in the sysroot at M0, for loopback only. Do not add TLS or download features. Verify licenses and font redistribution rights before packaging. Use CMake/CTest for host and cross builds; no dependencies fetched at runtime.

### Planned repository structure (create during implementation)

```text
README.md
docs/
  prd.md                     # current requirement baseline
  implementation-plan.md     # this plan
  compatibility.md           # later: exact tested device/firmware matrix
  launch-investigation.md    # later: M0 evidence and chosen contract
CMakeLists.txt
cmake/toolchains/            # KNULLI sysroot build definitions
src/
  main.cpp
  app/                       # state machine and orchestration
  discovery/                 # system definitions and gamelists
  library/                   # models, filters, random selection
  storage/                   # cache/settings/history
  platform/                  # KNULLI probe and launch adapter
  ui/                        # renderer, input, image lifecycle
assets/                      # licensed fonts and fallback art only
tests/
  fixtures/                  # synthetic configs/XML, no copyrighted ROMs
  integration/               # fake ES endpoint and file-tree fixtures
  unit/                      # public domain contracts
packaging/ports/              # wrapper and staged archive layout
packaging/app-store/          # submission metadata after verified release
tools/                       # reproducible build/package/benchmark tools
```

Do not create empty implementation directories in the planning change. Add files when a milestone uses them.

## 4. Discovery, interfaces, and data contracts

### System and metadata resolution

Implement the inspected ES precedence in the KNULLI adapter, with fixture tests and explicit firmware profiles if installed versions differ. Fixed **configuration locations** are allowed compatibility knowledge; ROM roots and emulator commands must never be baked in.

1. Select the first existing base configuration: user `es_systems_custom.cfg`, user `es_systems.cfg`, `/usr/share/emulationstation/es_systems.cfg`, then `/etc/emulationstation/es_systems.cfg`. The inspected user directory is `/userdata/system/configs/emulationstation`.
2. Apply `es_systems_*.cfg` additions from the user directory, then system directory. Matching system names replace supplied child fields; new names add systems. ES does not explicitly sort enumeration here. Record actual order and compare the effective result with ES; do not assume alphabetical or user-last behavior. Treat `es_systems_custom.cfg` consistently with the source's special handling.
3. Read system ID, display name, configured root, extensions, and launch capability. Expand supported ES path syntax using the installed version's rules. Never execute configuration strings as part of discovery.
4. Exclude virtual collections and the picker itself. Default to normal game systems; exclude utility Ports from All Systems to prevent recursive app launches. Document this support boundary and present unsupported systems diagnostically.
5. Prefer `<system-root>/gamelist.xml`; otherwise use the user ES directory's `gamelists/<system-id>/gamelist.xml`. Resolve game and metadata paths using ES rules, including root-relative and home-relative cases, not the XML file's directory. Verify artwork resolution against source fixtures.
6. Match ES scan behavior: normal mode scans configured extensions then overlays XML; `ParseGamelistOnly` and `IgnoreGamelist` alter this. Respect supported recursion settings, hidden/disabled entries, launchable directories, playlists and multi-disc references. Track visited directories to stop symlink loops. Do not index artwork, BIOS, folder nodes, or playlist member discs as separate games when ES suppresses them.
7. Read relevant ES recovery records after base metadata when their validity rule passes, or reconcile live metadata through ES. Never write or flush ES files. Test favorites changed in ES immediately before opening the Port.
8. Reconcile candidates with the loaded ES catalogue for an HTTP backend using `/systems` and `/systems/<id>/games`. Preserve exact ES launch paths separately from canonical deduplication paths. Unindexed/new ROMs must not appear launchable until ES knows them; explain that ES needs a library refresh. Do not trigger refreshes automatically.

Sources: [config selection](https://github.com/knulli-cfw/batocera-emulationstation/blob/10834d004ec3de0808703b46220b6a8574d53218/es-app/src/SystemData.cpp#L1332-L1347), [additions merge](https://github.com/knulli-cfw/batocera-emulationstation/blob/10834d004ec3de0808703b46220b6a8574d53218/es-app/src/SystemData.cpp#L806-L878), [gamelist precedence](https://github.com/knulli-cfw/batocera-emulationstation/blob/10834d004ec3de0808703b46220b6a8574d53218/es-app/src/SystemData.cpp#L1401-L1417), [gamelist/recovery parsing](https://github.com/knulli-cfw/batocera-emulationstation/blob/10834d004ec3de0808703b46220b6a8574d53218/es-app/src/Gamelist.cpp), [initial scan options](https://github.com/knulli-cfw/batocera-emulationstation/blob/10834d004ec3de0808703b46220b6a8574d53218/es-app/src/SystemData.cpp#L81-L130).

### Proposed data models

These are contracts, not committed C++ declarations. Use value types until a real shared boundary needs an interface.

| Model | Minimum fields and invariant |
|---|---|
| `SystemRecord` | Stable ES `id`, label, resolved ROM root, extensions, selected config/gamelist sources, capability status. An ID is not inferred from a directory name. |
| `GameRecord` | Stable key `(systemId, normalized root-relative path)`, canonical file identity, exact ES launch path, title, optional image/thumbnail/marquee paths, favorite/hidden flags, eligibility reason. Fallback title is filename stem. No decoded images or full descriptions in the index. |
| `LibrarySnapshot` | Generation, schema version, source fingerprints, systems, compact game vector, diagnostics, ES catalogue verification state. Publish a complete immutable generation, not half a scan. |
| `FilterSpec` | Optional system ID, favorites-only flag, recent-exclusion window (default disabled). Empty result is an explicit state. |
| `SelectionHistory` | Bounded ordered stable keys for successfully displayed selections; selection, not launch, defines “recent.” Removed games are harmless. |
| `Settings` | Schema version, last system, P1 flags/window, controller remap if needed, UI preferences. No emulator command or ROM-root defaults. Corrupt settings fall back with a visible notice. |
| `LaunchRequest` | Snapshot generation, stable game key, exact system ID and path. No shell command. Reject stale/missing/ambiguous candidates before dispatch. |
| `LaunchResult` | Not submitted (safe to retry), submitted (not proof of play), or outcome unknown. Include actionable reason. Never automatically retry unknown submissions. |

Public operations: `discover(capabilities, cancellation) -> LibrarySnapshot`; `load/saveSnapshot(sources)`; `eligible(snapshot, filter, history) -> vector<game-index>`; `pick(eligible, rng) -> optional<game-index>`; `probeLaunch() -> capabilities`; `submit(request) -> LaunchResult`. The UI receives worker results tagged with snapshot/request generation. It discards stale artwork responses and never retains pointers into replaced vectors.

### Selection contract

Choose once from a flat vector of eligible unique games with `std::uniform_int_distribution` over its full index range. Seed the C++ random engine once per process from operating-system entropy; fixed seeds are test-only. Do not choose a system first and do not use modulo reduction. All Systems gives each eligible title the same probability, so larger systems get proportionally more selections.

Collapse duplicate metadata entries and collection aliases. Distinct ROM files remain distinct games; no content hashing or region-title guessing in v1.0. A zero-game pool shows a filter/reset message; a one-game pool can repeat. P0 sampling is with replacement. P1 recent exclusion removes keys before drawing; when it empties the pool, ask the user to reset/disable exclusion rather than silently relax it. Favorites and recent filters intersect with the system filter. Surprise Mode uses the same eligibility and launch checks.

## 5. Launch investigation and lifecycle

### M0 experiment, before application development

On each exact target firmware, use a disposable test Port and legally supplied fixture games. Record commands, timestamps, logs, controller identity, return behavior, and video. Do not alter system emulator commands or global firmware configuration to make a test pass.

1. Inventory ES/configgen versions, executable locations, display backend, SDL versions, user config precedence, HTTP capabilities, and player mappings. Establish a normal ES launch control sample.
2. From the running test Port, probe loopback read-only endpoints with bounded timeouts. Check the intended path is present once in loaded real systems; compare system/core overrides with the normal ES control.
3. Release SDL/audio/input, send one raw-path `POST /launch`, then exit the wrapper promptly. Verify queue handling after outer configgen cleanup. Test whether safe ordering differs on each firmware; do not use an arbitrary sleep as the contract.
4. Verify exactly one correct game starts with built-in and external controllers, expected emulator/core, sound, scaling, hotkey exit, save behavior, and ES recent/playtime behavior. Exit back to a usable ES menu. Repeat transitions, not just the first launch.
5. Test endpoint absent, unmatched and duplicate paths, file removal, controller unplug, rapid double-press, HTTP timeout before/after send, selected-emulator failure, and cancellation before submission. No auto-resend after a possibly accepted request; no detached retry helper or replay of a saved pending request.
6. Compare direct configgen only if HTTP cannot satisfy the contract. Invoke the discovered KNULLI tool with an argv vector, not a shell; pass validated system/path and verified controller descriptors. Test nested evmapy, hooks, display/audio restoration, signals, and history attribution. A configgen fallback must not bypass a custom system command silently.
7. Record the chosen mechanism, supported versions, known errors, and trace evidence in `docs/launch-investigation.md`. If neither backend is safe, stop launch implementation and propose a small upstream ES handoff API. Do not replace the requirement with hardcoded emulator invocations.

No automatic backend fallback after submission: it can launch twice. Backend choice is made during capability validation, before the Play action. Freeze input once dispatch starts. Persist settings/history before releasing SDL. Failure before submission can restore the picker; once submitted, exit to ES. Since this API has no robust launch acknowledgement, describe dispatch as “submitted,” and measure successful play separately in hardware tests. A stale path must produce a clear diagnostic on the next picker open, not an endless pending launch.

## 6. Cache and performance plan

### Storage and invalidation

Use a versioned compact cache per system plus a small manifest. Keep settings/history as separate JSON files. Start with JSON cache records parsed one system at a time; change to a compact binary format or SQLite only if device profiles show it is needed. Store only fields used by filtering and preview. Cache is disposable; user settings are not.

- Resolve an app-owned writable state directory through the KNULLI adapter; a proposed location is `<user-system-dir>/configs/random-game-picker`. Obtain the user system directory from the validated firmware profile, not from a ROM-root assumption. Packaging must declare the final narrow write paths.
- Fingerprint effective config files, source ordering, gamelist/recovery files (path, size, modification time), scan options, firmware/adapter version, and root identity. Small config content hashes catch same-timestamp changes. ROM additions/removals need directory change detection, not just gamelist timestamps.
- Persist visited-directory fingerprints from each scan. Recheck them incrementally off the UI thread; additions change their parent directory. Offer a controller-accessible full Rescan for preserved timestamps, unreliable filesystem notifications, or replaced storage. Inotify is an optimization, not correctness infrastructure.
- Warm start loads a valid snapshot and filter vector first. Reconcile changed sources and the live ES catalogue in the worker. Mark cached records unverified until reconciliation; before Play, require current launch eligibility and file existence. Do not wait for a full recursive scan to draw the first screen.
- Changed systems rebuild independently. Cancellation and parse errors keep a usable old snapshot marked stale, but removed/unavailable roots cannot remain launchable. A malformed system does not discard other systems. Show skipped-system counts.
- Write a new cache to a temporary file in the same directory and atomically rename after completion. Bound file/record sizes and reject unknown schema versions. Corruption or interrupted writes cause rebuild, not crash. Store no executable command text in the cache.

### Time and memory budgets (proposed engineering gates)

| Operation | Design and measurement |
|---|---|
| Cached startup | Under 3 s from Port wrapper start to interactive default screen on each device, including settings/cache/input. Report median, p95, and worst of 30 runs; target all runs under 3 s on the reference card. Record cold OS-page-cache runs separately. |
| Selection/reroll | Under 200 ms from controller event to selected title/placeholder drawn; target p95 and worst under 200 ms over 1,000 actions. Includes filter application if changed; full image decode is asynchronous and reported separately. |
| Scaling | Test 10k, 25k, and 50k records. Initial discovery and filter rebuild are linear in entries; each draw from a built vector is constant time. Keep eligible vectors as indices, not record copies. No whole-library rescan per reroll. |
| Memory | Provisional peak RSS budget 96 MiB at 10k, including decoded image buffers; measure ES/system free memory too. Reassess with evidence before raising the budget. Track 25k/50k growth and low-memory behavior. |
| Artwork | Load selected art only, with a byte-bounded decoded/texture cache (initial target 16 MiB combined). Limit image dimensions/bytes before allocation; downsample large sources and preserve aspect ratio. Prefer image, then thumbnail, then marquee, then bundled fallback. Do not copy scraped assets. |
| Rendering | Target stable 30 fps on both devices; no roulette dependency. Decode in worker, create/destroy SDL textures on renderer thread; cancel obsolete requests on rapid rerolls. |

Cold scans show progress, permit Back/cancel, and do not promise a 3-second cold scan. A low-resolution device cannot retain full-screen 720p textures needlessly. Benchmark actual SD cards with realistic paths and metadata lengths, not only synthetic in-memory records.

## 7. UI, input, and device compatibility

States: Loading → Selector/Preview → Submitting → Exit; Y opens Settings/Filters from the selector. This follows the owner's 2026-10-19 flow update, superseding the original filters-first baseline. Empty library, empty filter, stale/unavailable launch, malformed metadata, and no controller are recoverable states, not modal dead ends. Background rebuilds keep focus stable.

- After discovery, select a game using saved system/favorites filters (All Systems on first use), without launching. Selector has “A: Play, X: Reroll, Y: Settings, B: Filters.” Settings retains system/favorites/recent controls and Y: Rescan; A on System selects and returns to preview. B from Filters exits; Start remains global quit. B never cancels a request already submitted. No extra launch confirmation in normal preview. Surprise Mode is explicit and off by default.
- Use a responsive logical layout with separate wide and 4:3 arrangements. Keep title, system, current filters, and action prompts visible at both sizes. Ellipsize or wrap long UTF-8 names; retain readable font sizes. Use artwork letterboxing and a labeled fallback. Disabled Play explains why.
- Use SDL GameController mappings inherited from KNULLI and available database entries. Translate physical A/B/X legends to application actions consistently with device layout. Provide a small remapping screen only where a verified mapping fails; keyboard is development-only.
- Handle D-pad/hat, stick dead zone and repeat, press/release debounce, hotplug, resume, and focus loss. Ignore held launch buttons during transition. Do not consume KNULLI power/hotkey controls or change global controller files.

| Target | Expected display | Required proof before support claim |
|---|---|---|
| TrimUI Smart Pro | 1280×720, wide layout | Exact KNULLI image, built-in controls, external pad, rendering backend, launch/return lifecycle, ABI and performance. Official device documentation exists. |
| MagicX Zero 28 | 640×480, 4:3 layout | Same proof plus exact model/device identity and firmware availability. Current source has a board target; the App Store reports a diagnostic with `magicx-input`, not full picker validation. Do not substitute another MagicX model. |

Build against the selected KNULLI sysroot, not the developer host's glibc. Inspect ELF architecture, required glibc symbols, dynamic dependencies, and SDL backend on each image. Prefer one aarch64 build only if both pass; otherwise publish separately identified artifacts. Do not bundle incompatible SDL drivers or replace firmware libraries. Verify KMS/DRM or other actual backend on hardware; a desktop window is not proof of handheld rendering.

## 8. Port and App Store delivery

First produce a manual Port archive with a launcher in the **configured Ports root** and an adjacent private application directory. Resolve the wrapper's own location robustly, set its working directory, check capabilities, and select the supported artifact. Do not overwrite ES configuration or assume `/userdata/roms/ports`. The install guide must explain how the user locates the configured Ports root. Keep runtime cache/settings outside replaceable binaries. No root-only install step, `curl | sh`, self-updater, ROM payload, or artwork download.

After device validation, follow [the App Store package guide](https://github.com/jellydn/knulli-app-store/blob/main/docs/how-to-add-a-package.md). Create a stable reverse-domain ID, license/maintainer/source fields, immutable versioned ZIP or tar.gz, exact compressed/extracted sizes and SHA-256, archive inventory, declared executable files, architecture/ABI/dependencies, exact firmware/device/display matrix, narrow install/write paths, and preserve paths. Inspect for traversal, absolute names, links, special files, and duplicate entries.

The store uses declared destinations; confirm how those map to the configured Ports location before submission. If its schema cannot represent the target installation safely, resolve that with the store maintainer rather than forcing a hardcoded ROM root. Test install, repeated install, repair, upgrade, failed-upgrade rollback, uninstall, and preservation of settings/history. Submit evidence and run that repository's current catalogue/tests; its Go commands are not this C++ project's build commands. Mark verified only for exact real-device evidence. App Store submission/publication is a separate approval step, not part of this planning task.

## 9. Phased milestones and exit gates

Estimates are engineering effort, not calendar promises. Hardware and firmware access are dependencies.

| Phase | Scope and deliverables | Exit gate | Estimate |
|---|---|---|---|
| M0 — integration proof | Device inventory, launch experiment, configuration/path fixtures, local-IPC interpretation, chosen backend and compatibility floor | Correct configured game starts once and returns cleanly on both devices; no unresolved lifecycle blocker. Owner accepts recorded offline interpretation and any fallback limitations. | 3–5 days plus hardware access |
| M1 — discovery/domain | CMake host/cross skeleton, config/XML parser, index, deduplication, filtering and uniform picker | AC01–03 fixtures pass; unscripted roots, unscripted extensions, unscripted system IDs work; 10k fixture scales without ROM writes. | 3–5 days |
| M2 — storage/performance | Cache generation/invalidation, JSON settings, worker cancellation, benchmark runner | Cached startup/selection/RSS budgets measured on both devices; stale/corrupt cache and deleted-ROM cases safe. | 2–4 days |
| M3 — playable P0 Port | SDL layout, input, async art, errors, proven launcher and manual archive | AC01–09 demonstrated; 100 rerolls without crash; two-action selection; launch matrix ≥95%, with no configuration-mismatch or duplicate-launch defect accepted. | 4–6 days |
| M4 — P1 | Favorites, recent selections, explicit Surprise Mode | Intersections, empty-pool behavior, persistence and immediate launch reuse the same eligibility/safety contract; no new backend. | 2–3 days |
| M5 — release/store | Clean install/update/removal tests, compatibility table, licenses, immutable archives/checksums, store manifest and evidence | All traceability rows closed, exact hardware matrix documented; approved submission to App Store. No untested device marked verified. | 2–4 days |

M3 can provide an early P0 preview; it is not evidence that M4/P1 is complete. If M0 fails, re-plan rather than continue UI work around an unproven launcher.

## 10. Testing and acceptance traceability

Use CTest with a small established C++ test framework selected at M1. Keep tests at module/public contracts, not private function calls. Use synthetic ROM files and metadata. Host sanitizers cover memory/undefined behavior; hardware runs cover SDL/driver and launcher behavior. CI builds host tests and cross artifacts but cannot claim hardware compatibility.

High-value fixtures: custom and additive system configs with asymmetric overrides; relocated roots and fallback gamelists; no gamelist; malformed XML in only one system; relative/absolute paths with spaces, apostrophes, shell metacharacters and non-ASCII text; hidden games; duplicate paths across systems; launchable folders; playlists; symlink loops; stale favorites recovery; deleted or unreadable roots; huge/corrupt images; interrupted cache writes; a 1-game system alongside a 99-game system. XML must not resolve external entities or access URLs. Bound metadata sizes and reject invalid launch strings such as embedded NUL bytes.

Random tests use fixed-seed reproducibility plus boundary inputs and a distribution diagnostic on the 1:99 library that would expose system-first sampling. Do not use a flaky statistical p-value as the only correctness proof. Empty/single pools, duplicate aliases, and recent/favorites intersections are deterministic tests. Production entropy is tested separately from deterministic fixtures.

| PRD criterion/metric | Verification and retained evidence | Phase |
|---|---|---|
| AC01 installed games | Compare indexed games to ES for multiple supported systems, with/without XML, external storage and deleted files; record exclusions and reasons. | M1/M3 |
| AC02 unbiased choice | Flat-pool contract tests, uneven 1:99 distribution diagnostic, duplicate suppression, fixed-seed edge cases. | M1 |
| AC03 console filter | Mixed-size systems and same-title games: only requested system selected; All Systems restores full pool. | M1 |
| AC04 artwork | Valid/missing/broken/oversized image fixtures; inspect device screenshots at both resolutions and rapid-reroll stale-art case. | M3 |
| AC05 reroll | 100 consecutive controller rerolls without restart/crash; track memory; exercise asynchronous art and scan completion. | M3 |
| AC06 configured emulator | On each device, at least 40 launches across at least 4 supported systems, including global/system/folder/game overrides, RetroArch and a configured standalone emulator where available. Compare to direct ES controls; record each result and exit behavior. Expand matrix where these cases need more samples. | M0/M3 |
| AC07 controls | Built-in and external controllers: cold start, all screens/errors, held buttons, hotplug, hotkey exit, return to ES; record interaction video. | M3 |
| AC08 offline/both devices | Disable external networking, use local files only; confirm no non-loopback socket attempts, accounts or downloads; repeat core flow on both exact images. Literal no-sockets interpretation must pass separately if chosen. | M0/M3 |
| AC09 Port install | Fresh manual install into resolved Ports root; discover entry, open/exit, repair/remove without ROM/save changes. Store lifecycle tests later. | M3/M5 |
| ≥95% successful launches | Report successes/attempts per device and overall with all failures classified. Count success only when correct game is playable with working input/audio and exits cleanly; HTTP 2xx is not success. Do not discard valid failed attempts to reach threshold. | M3/M5 |
| ≤2 controller actions | From main screen with default/retained valid filter, A selects preview in one action; second A launches. Filter setup is measured separately. | M3 |
| Cached <3 s; pick <200 ms | Timestamped 30-start and 1,000-action device benchmarks at 10k+; publish median/p95/worst, SD card and firmware. Missing artwork must not delay selection. | M2/M5 |
| Low memory, no ROM changes | RSS/CPU profiles at 10k/25k/50k; compare fixture-library checksums/permissions before and after app sessions. Writes limited to declared app paths, apart from normal ES/emulator launch writes. | M2/M5 |
| P1 favorites/recent/surprise | Persist settings, reflect ES favorites, exhaust pool without silent relaxation, and launch through identical validation/dispatch. | M4 |

The 95% target is not permission to ship known double-launch, wrong-emulator, data-loss, or stuck-control bugs. These are release blockers even if the aggregate rate passes.

## 11. Risks, mitigations, and remaining decisions

| Risk | Mitigation / release boundary |
|---|---|
| ES API absent, blocked while Port runs, or queue lost | M0 proof on installed versions; capability-gated support. Do not substitute arbitrary delays. Propose upstream support if both candidates fail. |
| API false success, timeout, or duplicate-path launches | Preflight loaded catalogue and file; block ambiguous paths; single submission; no blind retry or automatic backend switch. Record play success only on device. |
| Configgen nesting loses controller/configuration or corrupts restoration | Prefer ES handoff; fallback needs independent lifecycle matrix and explicit acceptance of limitations. |
| ES config/metadata behavior differs across releases | Pin supported firmware profiles and fixtures; reconcile with live ES; explain unsupported versions. Never change global ES settings. |
| Cached library is stale after card replacement or scraping | Root/source identity, background directory validation, launch-time checks, full Rescan, complete-generation swaps. |
| Large XML/art stalls or exhausts handheld memory | One-system parsing, bounded records/images, single worker and cancellable requests, memory/RSS budget. |
| Shared SoC hides different ABI/display/input behavior | Cross sysroot and two-device test matrix; no compatibility claim based on chip alone. |
| Favorites visible in ES but not flushed to XML | Recovery-aware metadata or live catalogue reconciliation; test change-before-Port case. |
| Store destination/schema does not fit resolved Ports root | Check schema before release packaging; agree safe integration with maintainer. Manual Port remains first delivery. |
| Untrusted paths/config/XML/images | Read-only discovery, no shell expansion/execution, argv for system tools, bounded parsing/decoding, archive inspection and narrow writes. Existing KNULLI user configuration remains a trusted firmware input, not app-generated code. |
| Exact device image/hardware unavailable | Continue host fixtures only; mark hardware milestones blocked, never claim both-device support from desktop tests. |

Before M0 implementation, obtain both devices (or trusted testers), identify exact images, and record approval of loopback IPC interpretation. Before M5, select the project license, stable package ID/maintainer, and final tested install/write paths. These are explicit decisions for implementation/release, not reasons to add application code to this planning change.
