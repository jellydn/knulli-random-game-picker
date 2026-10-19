# PRD baseline — KNULLI Random Game Picker

Version 1.0 · Proposed MVP · Baseline recorded 2026-10-19

Source: [original PRD and owner decisions](https://ampcode.com/threads/T-01a1519f-4420-70c9-a33c-c98e19fe2252). This is a structured requirements baseline, not a claim that the application exists. The [implementation plan](implementation-plan.md) separates proposals from verified upstream behavior.

## Product and scope

Reduce decision fatigue for retro handheld owners with large ROM collections. Build an open-source, standalone C++17/SDL2 application for KNULLI on TrimUI Smart Pro and MagicX Zero 28. Open it from Ports and later distribute it through the KNULLI App Store.

| ID | Priority | Requirement |
|---|---|---|
| F01 | P0 | Select a random installed game. |
| F02 | P0 | Launch it in one action using the configured emulator. |
| F03 | P0 | Select All Systems (default) or one console. |
| F04 | P0 | Show title, system, and existing box art when available. |
| F05 | P0 | Reroll without restarting. |
| F06 | P0 | Navigate with controllers; no touchscreen or keyboard required. |
| F07 | P1 | Select only favorites. |
| F08 | P1 | Exclude recently selected games. |
| F09 | P1 | Surprise Mode: select and launch without preview. |

Flow: open Port → select system → A to select → preview → A to play or X to reroll → B to change filters or exit. Use a dark retro-style theme at 1280×720 and 640×480. Roulette animation is optional. Missing artwork must have a useful fallback.

Owner flow update, 2026-10-19: open Port → automatically show a random game using saved filters → A to play or X to reroll. Y opens the existing settings/filters screen; Y there still rescans. B remains selector → filters → exit, and Start remains global quit. This supersedes the baseline's initial system-selection step. It is implemented on source `main`, not in published `0.1.0`; see [current controls](testing.md#controls).

## Constraints and non-functional requirements

- Discover games through EmulationStation system configuration and `gamelist.xml`; reuse scraped artwork.
- Do not hardcode ROM directories or emulator commands. Resolve launch behavior through KNULLI-compatible mechanisms.
- Store application settings in local JSON. Do not duplicate or modify ROMs.
- Work fully offline, without accounts, telemetry, or network access. Loopback IPC needs explicit interpretation; see plan decision Q1.
- Support 10,000+ games with low ARM memory use. Cached startup target: under 3 seconds. Random selection target: under 200 ms.
- Handle missing metadata and deleted ROMs without crashing.
- No ROM or artwork downloads, online recommendations, AI discovery, matchmaking, or cloud synchronization in v1.0.

## Acceptance criteria

| ID | Required outcome |
|---|---|
| AC01 | Detect installed games across supported systems. |
| AC02 | Select games without favoring particular systems or titles. |
| AC03 | Filter by console. |
| AC04 | Display available artwork. |
| AC05 | Reroll without restarting. |
| AC06 | Launch with the correct emulator configuration. |
| AC07 | Navigate entirely with handheld controls. |
| AC08 | Work offline on both target devices. |
| AC09 | Install and start through Ports. |

Success metrics: at least 95% successful launches among supported, correctly configured games; random selection within two controller actions from the main screen; no crashes during 100 consecutive rerolls.

## Open questions (original wording)

1. Can the existing KNULLI launcher be invoked reliably from a standalone SDL2 application?
2. How should the application handle multiple emulator configurations for one system?
3. Should the picker launch as a Port or integrate directly into EmulationStation?

Owner decisions: use a new personal repository under `dunghd`, named `knulli-random-game-picker`; prepare an implementation plan, not application code.
