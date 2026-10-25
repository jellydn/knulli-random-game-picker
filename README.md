# KNULLI Random Game Picker

A Go and gabagool/v2 application that selects games from an existing KNULLI library. Target devices are the TrimUI Smart Pro and MagicX Zero 28.

**Version `0.2.1`: Go/gabagool maintenance release, KNULLI only.** Uses the selected
centered-artwork console layout, gabagool settings, separate discovery/launcher,
firmware/input profiles and typed picker state. No other CFW support is claimed.
See [architecture and research](docs/go-rewrite.md) for scope and extension rules.
Requires firmware SDL2/image/ttf/**gfx** and aarch64 glibc **2.34+**.

Earlier 0.2.0 TSP development-build testing verified dependency resolution, selector/settings,
injected controller events under both effective layouts, 100 paced rerolls,
exact-path emulator handoff and ES display restoration. Physical buttons,
audible audio, normal gameplay/hotkey return and MagicX remain unverified.
See the [0.2.0 release notes](docs/release-0.2.0.md) for evidence and limits.
The existing `0.2.0` and earlier releases and their assets are unchanged.
See the [0.2.1 release notes](docs/release-0.2.1.md) for this maintenance release;
no new hardware validation is claimed.

This release includes system/favorites filters, uniform random selection, preview of existing PNG/JPEG scraped artwork, reroll, session recent exclusion, controller navigation, and a single-submit ES launch handoff. It reads system configurations and gamelists, reconciles with the running ES catalogue, and never builds an emulator command or guesses a ROM directory.

**New in `0.1.1`:** the game selector opens after discovery, using saved filters. A plays the selected game; Y opens settings, where Y still rescans. Saved system selection is no longer reset before discovery. Earlier `0.1.0` opens filters first; its published assets remain unchanged.

**New in `0.2.1`:** the selector redraws on events/results rather than on
every idle poll; discovery reuses resolved paths within each scan, never across
Play verification. Invalid settings fall back as a whole. Packaging enforces
GLIBC ≤2.34 even with Python optimization and records build input versions.
The centered-artwork UI, both controller layouts and exact-path single-submit
launch contract are unchanged. No public archive or tag was replaced.

There is still no persistent library cache or Surprise Mode. Artwork decoding
is bounded but synchronous; host benchmarks do not prove device budgets.
Remaining hands-on checks are tracked in [TSP controls/audio/return #6](https://github.com/jellydn/knulli-random-game-picker/issues/6),
[overrides/second system #7](https://github.com/jellydn/knulli-random-game-picker/issues/7),
[MagicX #8](https://github.com/jellydn/knulli-random-game-picker/issues/8), and
[precise device performance #9](https://github.com/jellydn/knulli-random-game-picker/issues/9).
See the [audited codemap](.planning/codebase/CONCERNS.md) for each concern's status.

- [Download releases](https://github.com/jellydn/knulli-random-game-picker/releases).
- [Installation and testing](docs/testing.md): requirements, controls, logs, known limits, and the hardware report checklist.
- [Development setup and CI](docs/development.md): host build/tests and AArch64 packaging.

- [Implementation plan](docs/implementation-plan.md): architecture, contracts, milestones, test gates, risks, and PRD traceability.
- [PRD baseline](docs/prd.md): requirements and open questions from the source conversation.

MIT licensed; third-party components have their own [notices](THIRD_PARTY.md). No ROMs, commercial artwork, network recommendations, or telemetry are included.
