# Development setup and CI

The repository is still planning-only. Setup prepares a **host** C++17/SDL2 development environment. It does not implement the picker, create a CMake application, install a Port, or prove KNULLI compatibility.

## Orb setup

Amp runs executable `.agents/setup` in a fresh orb after the file reaches the project's default branch. It installs missing Debian/Ubuntu packages for C++, CMake, Ninja, pkg-config, SDL2/image/ttf, pugixml, nlohmann/json, Python/PyYAML, and ShellCheck. It skips package installation on warm runs. Run it manually with:

```sh
./.agents/setup
./scripts/check
```

No credentials, services, or environment-file changes are required. There is no resume hook because there is nothing to reconnect. Packages come from the host distribution; this is not a pinned KNULLI sysroot. Device ABI and cross-build setup remain M0/M1 work in the [implementation plan](implementation-plan.md).

## Checks

`scripts/check` is the shared local/CI command. It performs:

- Bash syntax and ShellCheck checks.
- Local Markdown file-link, whitespace, code-fence, and conflict-marker checks.
- Required PRD acceptance, open-question, and milestone marker checks.
- Workflow YAML parsing and full-commit action-pin checks. This is not a full GitHub Actions schema validator.
- A temporary C++17 compile/link/load probe for the selected host dependencies. It opens no window and uses no controller or network. Temporary files are removed on exit.

There are no product tests, application builds, release assets, or hardware results yet. These checks must not be described as proof of game discovery, rendering, controller support, or launching.

## GitHub Actions

`.github/workflows/check.yml` runs on pull requests, pushes to `main`, and manual dispatch. It uses read-only repository permissions, disables persisted checkout credentials, limits runtime, and cancels obsolete checks on the same ref. It calls the same setup and check commands as local development.

This follows the [KNULLI App Store setup](https://github.com/jellydn/knulli-app-store/blob/main/.agents/setup) and [check workflow](https://github.com/jellydn/knulli-app-store/blob/main/.github/workflows/check.yml) patterns. Its [original setup/CI thread](https://ampcode.com/threads/T-01a0a370-234b-729a-9b63-3871aa36a52c) records the method: inspect the environment, add repeatable setup, use shared checks, and verify locally before delivery. The checkout action pin matches the inspected App Store workflow. Go builds, catalogue generation, and release publication do not apply here.

The current repository remote is Amp-hosted. GitHub does not execute these files there. A GitHub repository/mirror and a push are required to activate Actions. Do not create that repository, publish code, or trigger a release without owner approval. Once application code exists, add real CMake/CTest, cross-build, and device evidence jobs in their planned milestones; do not use dummy product builds now.
