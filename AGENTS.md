# Project: spm (Smart Package Manager)

Go CLI that auto-detects npm/yarn/pnpm and proxies commands.

## Development

```bash
just setup               # install git hooks and dev tools
just test                # run tests
just fmt                 # format all Go files
just build               # build binary
```

## Conventions

- **Changelog**: Always update `CHANGELOG.md` when making user-facing changes. Add entries under `[Unreleased]` using the appropriate category (Added, Changed, Deprecated, Removed, Fixed, Security). Follow [Keep a Changelog](https://keepachangelog.com/en/1.0.0/) format.
- **README**: Keep `README.md` up to date when adding new commands, flags, or changing behavior.
- **Tests**: Add or update tests for any new functionality in the corresponding `_test.go` files.
- **Formatting**: A pre-commit hook runs `goimports` on staged Go files. Run `just setup` after cloning.
- **CI**: GitHub Actions runs `go test ./... -v -race`, format checks, and cross-platform builds on every PR.
