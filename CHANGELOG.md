# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

<!-- NOTE FOR AGENTS: When making changes to this project, always update the
     [Unreleased] section below with a summary of your changes, using the
     appropriate category (Added, Changed, Deprecated, Removed, Fixed, Security).
     When a release is cut, move Unreleased entries under a new version heading. -->

## [Unreleased]

### Added

- Type-to-filter support on interactive script selection (`spm run`).

### Fixed

- Include CHANGELOG release notes in GitHub releases instead of auto-generated content.

## [0.6.1] - 2026-03-30

### Added

- Gradient fade on install log lines — older lines appear more faded, most recent line at normal dim brightness. Increased visible log lines from 3 to 5.

### Fixed

- Fixed terminal escape sequence leaks (e.g., `^[[?2026;2$y`) after TUI exit when subprocesses finish quickly.

## [0.6.0] - 2026-03-27

### Added

- Interactive package search TUI — `spm add` with no arguments launches a type-ahead search against the npm registry.
- Install progress TUI — `spm install` now shows a spinner with elapsed time and scrolling log lines (last 3 lines of PM output) by default on TTY environments.
- `--raw` flag on `spm install` to bypass the progress TUI and display raw package manager output (auto-enabled in non-TTY/CI environments).

## [0.5.0] - 2026-03-26

### Added

- `spm init [npm|yarn|pnpm|bun]` command — initialize a new project with the chosen package manager, with interactive selection when no PM is specified.
- Homebrew tap distribution — install via `brew install decampsrenan/tap/spm`.
- `spm audit` command — runs a security audit and normalizes output across npm, yarn (classic & Berry), and pnpm.
- `--prod-only` flag on `spm audit` to skip dev dependencies.
- `--json` flag on `spm audit` for machine-readable output.
- `--severity` flag on `spm audit` to filter by minimum severity level (info, low, moderate, high, critical).
- `spm upgrade` command — self-update spm to the latest GitHub release, with `--alpha` for pre-releases, `--force` to reinstall, and `--dry-run` support.

## [0.4.0] - 2026-03-20

### Added

- Bun package manager support — detects `bun.lock` and legacy `bun.lockb` lock files, with deduplication when both are present.
- Colorized terminal output using Charmbracelet lipgloss — success, error, warning, info, and dim styles throughout the CLI.
- New `internal/ui` package with centralized style definitions and output helpers.
- Automatic alpha releases published on every push to `main` with unreleased changes.
- `--alpha` flag on install script to install the latest pre-release version.

### Fixed

- Kill vibes music subprocess when parent process dies (orphan detection via PPID polling).
- Handle SIGTERM in runner process, exiting with code 143.

### Changed

- Replaced `survey/v2` interactive prompts with Charmbracelet `huh/v2` for a modern TUI look.
- Styled `spm clean` output with colored paths, success checkmarks, and dimmed dry-run notices.
- Styled `--dry-run` command preview with violet highlighting.

## [0.3.0] - 2026-03-16

### Added

- `spm run` command with interactive script selection from package.json when no script is specified.
- Interactive prompt to select a package manager when `package.json` exists but no lock file is found.
- `spm remove <package>` command to uninstall packages, translating to the correct command for each package manager.
- `spm clean` command to remove `node_modules` and optionally the lock file (`--lock`).
- `--yes` / `-y` flag on `spm clean` to skip the confirmation prompt (useful in CI).

### Changed

- Revamped README with modern layout, badges, friendly tone, and collapsible table of contents.
- Extracted contributing instructions into a dedicated `CONTRIBUTING.md` file.
- `spm add` and `spm remove` now show contextual error messages when called without packages.
- `spm clean` now prints "Nothing to remove." when targets don't exist, instead of prompting for removal.

### Fixed

- `spm install` now correctly passes through unknown flags (e.g., `--legacy-peer-deps`, `--frozen-lockfile`) to the underlying package manager.
- `spm clean` no longer fails when `package.json` exists but no lock file is found.
- `spm clean` no longer falsely prints "Removed node_modules" when the directory did not exist.

## [0.2.1] - 2026-03-13

### Fixed

- Background music now stops when the user presses Ctrl+C (SIGINT) during command execution.

## [0.2.0] - 2026-03-13

### Added

- `--notify` flag to play a completion sound when commands finish (success/error).
- Notification sound files: `notification-pop.mp3`, `error-001.mp3`, `ding.mp3`.

### Changed

- Unified audio handling into a single implementation using beep with purego, removing CGO build-tag stubs.
- Audio playback (vibes music and notifications) now runs in detached child processes for non-blocking CLI usage.
- Release workflow is now manually triggered via `workflow_dispatch` with semantic version bump selection (patch/minor/major).
- Changelog is committed directly to main during release instead of creating a separate PR.
- Reinforced agent instructions in AGENTS.md to enforce changelog maintenance on feature branches.

### Fixed

- `--vibes` flag now works correctly with `CGO_ENABLED=0` builds on macOS.

## [0.1.5] - 2026-03-13

### Added

- `--vibes` flag to play background music during `spm install`.

### Fixed

- `--dry-run` flag now works correctly when placed before the script name (e.g., `spm --dry-run dev`).

### Added

- Pre-commit hook to auto-format Go code and organize imports via `goimports`.
- `justfile` with `setup`, `fmt`, `test`, and `build` recipes.
- CI format check job to catch unformatted code.
- Tests for `cmd`, `runner`, and `prompt` packages.
- Fallback to `~/.local/bin` in install script when `/usr/local/bin` is not writable.
- PATH warning when install directory is not in user's PATH.
- `--version` / `-v` flag to display the build version.
- `just dev` recipe using `gow` for watch-mode rebuilds during development.

## [0.1.0] - 2026-03-11

### Added

- Auto-detection of package manager (npm, yarn, pnpm) via lock files.
- Directory walk-up from current directory to `$HOME` to find the project root.
- Interactive prompt when multiple lock files are detected.
- `spm install` / `spm i` command to install dependencies.
- `spm add <packages...>` command to add packages.
- Fallback script execution (e.g., `spm dev` → `yarn dev` / `npm run dev`).
- `--dry-run` flag to preview the resolved command without executing it.
- Cross-platform build configuration via GoReleaser (linux/darwin × amd64/arm64).
- Curl-based installation script (`scripts/install.sh`).

[Unreleased]: https://github.com/DecampsRenan/spm/compare/v0.6.1...HEAD
[0.6.1]: https://github.com/DecampsRenan/spm/compare/v0.6.0...v0.6.1
[0.6.0]: https://github.com/DecampsRenan/spm/compare/v0.5.0...v0.6.0
[0.5.0]: https://github.com/DecampsRenan/spm/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/DecampsRenan/spm/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/DecampsRenan/spm/compare/v0.2.1...v0.3.0
[0.2.1]: https://github.com/DecampsRenan/spm/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/DecampsRenan/spm/compare/v0.1.5...v0.2.0
[0.1.5]: https://github.com/DecampsRenan/spm/compare/v0.1.0...v0.1.5
[0.1.0]: https://github.com/DecampsRenan/spm/releases/tag/v0.1.0
