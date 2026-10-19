# KNULLI Random Game Picker

A C++17 and SDL2 application that selects games from an existing KNULLI library. Target devices are the TrimUI Smart Pro and MagicX Zero 28.

**Status: experimental `0.1.0-alpha.2`, for owner-authorized device testing.** This fixes the reversed face-button actions reported on TrimUI Smart Pro in alpha.1; hardware retest is needed. No handheld has passed the full verification matrix. Desktop tests and ELF inspection do not establish real KNULLI launch compatibility. Do not treat this as the completed PRD MVP.

The first alpha includes system/favorites filters, uniform random selection, preview of existing PNG/JPEG scraped artwork, reroll, session recent exclusion, controller navigation, and an experimental single-submit ES launch handoff. It reads system configurations and gamelists, reconciles with the running ES catalogue, and never builds an emulator command or guesses a ROM directory.

- [Download prereleases](https://github.com/jellydn/knulli-random-game-picker/releases).
- [Installation and testing](docs/testing.md): requirements, controls, logs, known limits, and the hardware report checklist.
- [Development setup and CI](docs/development.md): host build/tests and AArch64 packaging.

- [Implementation plan](docs/implementation-plan.md): architecture, contracts, milestones, test gates, risks, and PRD traceability.
- [PRD baseline](docs/prd.md): requirements and open questions from the source conversation.

MIT licensed; third-party components have their own [notices](THIRD_PARTY.md). No ROMs, commercial artwork, network recommendations, or telemetry are included.
