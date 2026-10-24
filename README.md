# KNULLI Random Game Picker

A Go and gabagool/v2 application that selects games from an existing KNULLI library. Target devices are the TrimUI Smart Pro and MagicX Zero 28.

**Version `0.2.0`: Go/gabagool rewrite, KNULLI only.** Uses the selected
centered-artwork console layout, gabagool settings, separate discovery/launcher,
firmware/input profiles and typed picker state. No other CFW support is claimed.
See [architecture and research](docs/go-rewrite.md) for scope and extension rules.
Requires firmware SDL2/image/ttf/**gfx** and aarch64 glibc **2.34+**.

TSP development-build testing verified dependency resolution, selector/settings,
injected controller events under both effective layouts, 100 paced rerolls,
exact-path emulator handoff and ES display restoration. Physical buttons,
audible audio, normal gameplay/hotkey return and MagicX remain unverified.
See the [0.2.0 release notes](docs/release-0.2.0.md) for evidence and limits.
The existing `0.1.1` C++ release and its assets are unchanged.

This release includes system/favorites filters, uniform random selection, preview of existing PNG/JPEG scraped artwork, reroll, session recent exclusion, controller navigation, and a single-submit ES launch handoff. It reads system configurations and gamelists, reconciles with the running ES catalogue, and never builds an emulator command or guesses a ROM directory.

**New in `0.1.1`:** the game selector opens after discovery, using saved filters. A plays the selected game; Y opens settings, where Y still rescans. Saved system selection is no longer reset before discovery. Earlier `0.1.0` opens filters first; its published assets remain unchanged.

- [Download releases](https://github.com/jellydn/knulli-random-game-picker/releases).
- [Installation and testing](docs/testing.md): requirements, controls, logs, known limits, and the hardware report checklist.
- [Development setup and CI](docs/development.md): host build/tests and AArch64 packaging.

- [Implementation plan](docs/implementation-plan.md): architecture, contracts, milestones, test gates, risks, and PRD traceability.
- [PRD baseline](docs/prd.md): requirements and open questions from the source conversation.

MIT licensed; third-party components have their own [notices](THIRD_PARTY.md). No ROMs, commercial artwork, network recommendations, or telemetry are included.
