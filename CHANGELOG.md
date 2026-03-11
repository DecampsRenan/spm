# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

<!-- NOTE FOR AGENTS: When making changes to this project, always update the
     [Unreleased] section below with a summary of your changes, using the
     appropriate category (Added, Changed, Deprecated, Removed, Fixed, Security).
     When a release is cut, move Unreleased entries under a new version heading. -->

## [Unreleased]

### Fixed

- `--dry-run` flag now works correctly when placed before the script name (e.g., `spm --dry-run dev`).

### Added

- Pre-commit hook to auto-format Go code and organize imports via `goimports`.
- `justfile` with `setup`, `fmt`, `test`, and `build` recipes.
- CI format check job to catch unformatted code.
- Tests for `cmd`, `runner`, and `prompt` packages.
- Fallback to `~/.local/bin` in install script when `/usr/local/bin` is not writable.
- PATH warning when install directory is not in user's PATH.

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

[Unreleased]: https://github.com/DecampsRenan/spm/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/DecampsRenan/spm/releases/tag/v0.1.0
