# Project: spm (Smart Package Manager)

Go CLI that auto-detects npm/yarn/pnpm and proxies commands.

## Development

```bash
go test ./... -v -race   # run tests
go build -o spm .        # build
```

## Conventions

- **Changelog**: Always update `CHANGELOG.md` when making user-facing changes. Add entries under `[Unreleased]` using the appropriate category (Added, Changed, Deprecated, Removed, Fixed, Security). Follow [Keep a Changelog](https://keepachangelog.com/en/1.0.0/) format.
- **README**: Keep `README.md` up to date when adding new commands, flags, or changing behavior.
- **Tests**: Add or update tests for any new functionality in the corresponding `_test.go` files.
- **CI**: GitHub Actions runs `go test ./... -v -race` and cross-platform builds on every PR.
