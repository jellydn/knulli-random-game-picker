# KNULLI Random Game Picker

A C++17 and SDL2 application that selects games from an existing KNULLI library. Target devices are the TrimUI Smart Pro and MagicX Zero 28.

**Status: stable `0.1.0`, the first stable release of the current feature set.** The owner confirmed that the corrected alpha.2 works on TrimUI Smart Pro. MagicX Zero 28 and the full hardware launch/return/performance matrix remain unverified. This release does not complete the full PRD MVP.

This release includes system/favorites filters, uniform random selection, preview of existing PNG/JPEG scraped artwork, reroll, session recent exclusion, controller navigation, and a single-submit ES launch handoff. It reads system configurations and gamelists, reconciles with the running ES catalogue, and never builds an emulator command or guesses a ROM directory.

**Unreleased source change:** `main` opens the game selector after discovery, using saved filters. A plays the selected game; Y opens settings. Published `0.1.0` still opens filters first. No new release or App Store update is included in this change.

- [Download releases](https://github.com/jellydn/knulli-random-game-picker/releases).
- [Installation and testing](docs/testing.md): requirements, controls, logs, known limits, and the hardware report checklist.
- [Development setup and CI](docs/development.md): host build/tests and AArch64 packaging.

- [Implementation plan](docs/implementation-plan.md): architecture, contracts, milestones, test gates, risks, and PRD traceability.
- [PRD baseline](docs/prd.md): requirements and open questions from the source conversation.

MIT licensed; third-party components have their own [notices](THIRD_PARTY.md). No ROMs, commercial artwork, network recommendations, or telemetry are included.
