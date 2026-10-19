# KNULLI Random Game Picker

A C++17 and SDL2 application that selects games from an existing KNULLI library. Target devices are the TrimUI Smart Pro and MagicX Zero 28.

**Status: stable `0.1.1`, with selector-first startup.** The owner confirmed that the earlier corrected alpha.2 works on TrimUI Smart Pro; the new startup flow still needs handheld retesting. MagicX Zero 28 and the full hardware launch/return/performance matrix remain unverified. This release does not complete the full PRD MVP.

This release includes system/favorites filters, uniform random selection, preview of existing PNG/JPEG scraped artwork, reroll, session recent exclusion, controller navigation, and a single-submit ES launch handoff. It reads system configurations and gamelists, reconciles with the running ES catalogue, and never builds an emulator command or guesses a ROM directory.

**New in `0.1.1`:** the game selector opens after discovery, using saved filters. A plays the selected game; Y opens settings, where Y still rescans. Saved system selection is no longer reset before discovery. Earlier `0.1.0` opens filters first; its published assets remain unchanged.

- [Download releases](https://github.com/jellydn/knulli-random-game-picker/releases).
- [Installation and testing](docs/testing.md): requirements, controls, logs, known limits, and the hardware report checklist.
- [Development setup and CI](docs/development.md): host build/tests and AArch64 packaging.

- [Implementation plan](docs/implementation-plan.md): architecture, contracts, milestones, test gates, risks, and PRD traceability.
- [PRD baseline](docs/prd.md): requirements and open questions from the source conversation.

MIT licensed; third-party components have their own [notices](THIRD_PARTY.md). No ROMs, commercial artwork, network recommendations, or telemetry are included.
