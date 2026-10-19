# KNULLI Random Game Picker

A planned C++17 and SDL2 application that selects and launches games from an existing KNULLI library. Target devices are the TrimUI Smart Pro and MagicX Zero 28.

**Status: planning only.** This repository does not contain an application, launcher, build system, or release package yet.

- [Implementation plan](docs/implementation-plan.md): architecture, contracts, milestones, test gates, risks, and PRD traceability.
- [PRD baseline](docs/prd.md): requirements and open questions from the source conversation.
- [Development setup and CI](docs/development.md): orb toolchain, shared checks, and GitHub Actions activation.

The first implementation milestone must prove a safe launch handoff from a running Port to EmulationStation on both devices. Source inspection is not a substitute for this device test.
