# Contributing to spm

First off, thanks for taking the time to contribute! 🎉

All types of contributions are encouraged and valued — whether it's a bug report, a feature request, or a pull request. I maintain this project on my own, so any help is appreciated. See the sections below for more details.

## How Can I Contribute?

### Reporting Bugs

If you find a bug, please [open an issue](https://github.com/DecampsRenan/spm/issues/new?labels=bug&title=bug%3A+) and include:

- A clear and descriptive title
- Steps to reproduce the behavior
- What you expected to happen
- What actually happened
- Your environment (OS, Go version, spm version)

### Suggesting Features

Have an idea? [Open a feature request](https://github.com/DecampsRenan/spm/issues/new?labels=enhancement&title=feat%3A+) and describe:

- The problem you're trying to solve
- How you'd like it to work
- Any alternatives you've considered

### Pull Requests

1. Fork the repo and create your branch from `main`
2. If you've added functionality, add tests
3. Make sure the test suite passes
4. Make sure your code follows the existing style (run `just fmt`)
5. Update `README.md` if you changed any user-facing behavior
6. Update `CHANGELOG.md` with an entry under `[Unreleased]`
7. Open your pull request!

## Development Setup

### Prerequisites

- [Go](https://go.dev/) 1.25+
- [just](https://github.com/casey/just) — command runner

### Getting Started

```sh
# Clone the repository
git clone https://github.com/decampsrenan/spm.git
cd spm

# Install dev tools and git hooks
just setup

# Run tests
just test

# Format code
just fmt

# Build
just build

# Watch and rebuild on changes
just dev
```

### Project Structure

```
cmd/           # CLI commands (Cobra)
internal/
  audio/       # Audio playback (vibes & notifications)
  detector/    # Package manager detection
  prompt/      # Interactive prompts
  resolver/    # Command resolution
  runner/      # Command execution
```

## Style Guide

- Run `just fmt` before committing — a pre-commit hook enforces this
- Write tests in `_test.go` files alongside the code they test
- Keep commits focused and use clear commit messages
