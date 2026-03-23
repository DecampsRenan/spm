# Project: spm (Smart Package Manager)

Go CLI that auto-detects npm/yarn/pnpm/bun and proxies commands.

**Language rule**: This is an open-source project. All code, comments, commit messages, PR descriptions, documentation, and agent output **must** be written in English.

## Development

```bash
just setup               # install git hooks and dev tools
just test                # run tests
just fmt                 # format all Go files
just build               # build binary
```

## Conventions

- **README (REQUIRED)**: When adding or changing flags, commands, or user-facing behavior, you **must** update `README.md` (features list + usage examples). The CI reviewer will flag this.
- **Tests**: Add or update tests for any new functionality in the corresponding `_test.go` files.
- **Dry-run (REQUIRED)**: Any new functionality that executes commands or produces side effects must respect the `--dry-run` flag. Ensure `runner.Run` (or any new execution path) short-circuits correctly when `dryRun` is `true`. Tests must cover the dry-run case.
- **Formatting**: A pre-commit hook runs `goimports` on staged Go files. Run `just setup` after cloning.
- **CI**: GitHub Actions runs `go test ./... -v -race`, format checks, and cross-platform builds on every PR.

## Changelog (REQUIRED)

**Every code change MUST include an update to the CHANGELOG.**

Before submitting or pushing changes:
1. Open `CHANGELOG.md`
2. Add an entry under `## [Unreleased]` in the appropriate category:
   - **Added**: new features
   - **Changed**: changes to existing features
   - **Deprecated**: features that will be removed soon
   - **Removed**: features that have been removed
   - **Fixed**: bug fixes
   - **Security**: vulnerability fixes
3. CI **will block the merge** if no entry is present in `[Unreleased]`
4. Never modify versioned sections (e.g. `[0.1.5]`), only `[Unreleased]`

## Release

Releases are triggered **manually** via GitHub Actions:
1. Go to **Actions > Release > Run workflow**
2. Choose the bump type: `patch`, `minor`, or `major`
3. The workflow computes the version, updates the changelog, tags, and publishes via GoReleaser

**Alpha releases** are published automatically on every push to `main` that has entries in `[Unreleased]`. They are cleaned up when a stable release is published.
