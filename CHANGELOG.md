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

- `--notify` flag to play a completion sound when commands finish (success/error).
- Notification sound files: `notification-pop.mp3`, `error-001.mp3`, `ding.mp3`.

### Changed

- Unified audio handling into a single implementation using beep with purego, removing CGO build-tag stubs.
- Audio playback (vibes music and notifications) now runs in detached child processes for non-blocking CLI usage.

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

[Unreleased]: https://github.com/DecampsRenan/spm/compare/v0.1.5...HEAD
[0.1.5]: https://github.com/DecampsRenan/spm/compare/v0.1.0...v0.1.5
[0.1.0]: https://github.com/DecampsRenan/spm/releases/tag/v0.1.0
