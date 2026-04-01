<!-- NOTE FOR AGENTS: Keep this README up to date when adding new commands,
     flags, or changing behavior. Update the usage examples and feature list. -->

<h1 align="center">
  📦 spm
</h1>

<div align="center">
  <strong>Smart Package Manager</strong> — One CLI to rule them all. Stop worrying about npm vs yarn vs pnpm vs bun vs deno, just run <code>spm</code> and let it figure out the rest.
  <br />
  <br />
  <a href="https://github.com/DecampsRenan/spm/issues/new?labels=bug&title=bug%3A+">🐛 Report a Bug</a>
  ·
  <a href="https://github.com/DecampsRenan/spm/issues/new?labels=enhancement&title=feat%3A+">✨ Request a Feature</a>
</div>

<div align="center">
<br />

[![GitHub release](https://img.shields.io/github/v/release/DecampsRenan/spm?style=flat-square)](https://github.com/DecampsRenan/spm/releases) [![CI](https://img.shields.io/github/actions/workflow/status/DecampsRenan/spm/ci.yml?branch=main&style=flat-square)](https://github.com/DecampsRenan/spm/actions/workflows/ci.yml) [![License](https://img.shields.io/github/license/DecampsRenan/spm?style=flat-square)](LICENSE)

</div>

<details open="open">
<summary>Table of Contents</summary>

- [About](#about)
  - [Built With](#built-with)
- [Getting Started](#getting-started)
  - [Homebrew (macOS / Linux)](#homebrew-macos--linux)
  - [Install script (macOS / Linux)](#install-script-macos--linux)
  - [From source](#from-source)
- [Usage](#usage)
  - [Command mapping](#command-mapping)
- [Contributing](#contributing)
- [Support](#support)
- [License](#license)
- [Acknowledgements](#acknowledgements)

</details>

---

## About

Ever joined a project and had to check which package manager it uses before running anything? Yeah, me too. 😅

**spm** detects your project's package manager automatically and translates your commands on the fly. Just type `spm install`, `spm add react`, or `spm dev` — it handles the rest.

- 🔍 **Auto-detection** via lock files (`package-lock.json`, `yarn.lock`, `pnpm-lock.yaml`, `bun.lock`, `deno.lock`) and project markers (`package.json`, `deno.json`, `deno.jsonc`)
- 📂 **Directory walk-up** — works from any subdirectory in your project
- 🔀 **Flag pass-through** — unknown flags (e.g. `--legacy-peer-deps`) are forwarded to the underlying package manager
- 🔄 **Command translation** — maps commands to the correct syntax for each package manager
- 🎯 **Interactive script runner** — `spm run` lets you pick a script from package.json (or a task from deno.json)
- 💬 **Interactive prompt** when multiple lock files are detected or no lock file exists
- 👀 **Dry-run mode** to preview commands without executing them
- 🎵 **Vibes mode** — play background music while installing dependencies (`--vibes`)
- 🔔 **Notification sounds** — get notified when the command finishes (`--notify`)
- 🆕 **Project init** — `spm init` scaffolds a new project with the package manager of your choice
- 🔒 **Security audit** — `spm audit` runs a dependency audit and normalizes output across npm/yarn/pnpm
- ⬆️ **Self-upgrade** — `spm upgrade` updates spm to the latest release from GitHub
- 🔎 **Interactive search** — `spm add` with no args lets you search and pick a package from the npm registry
- 📋 **Package details** — press Enter on a search result to view metadata (license, downloads, stars, repo) and pick a specific version to install
- ✨ **Progress TUI** — `spm install` shows a live spinner with elapsed time and scrolling logs (use `--raw` for raw output)

### Built With

- [Go](https://go.dev/)
- [Cobra](https://github.com/spf13/cobra)
- [Charmbracelet](https://github.com/charmbracelet) (lipgloss, huh, bubbletea)
- [Beep](https://github.com/gopxl/beep)

## Getting Started

### Homebrew (macOS / Linux)

```sh
brew install decampsrenan/tap/spm
```

### Install script (macOS / Linux)

The fastest way to get started:

```sh
curl -fsSL https://raw.githubusercontent.com/decampsrenan/spm/main/scripts/install.sh | bash
```

Want to install to a custom directory? No problem:

```sh
INSTALL_DIR=~/.local/bin curl -fsSL https://raw.githubusercontent.com/decampsrenan/spm/main/scripts/install.sh | bash
```

### Alpha builds

Want to try the latest unreleased features? Install the most recent alpha:

```sh
curl -fsSL https://raw.githubusercontent.com/decampsrenan/spm/main/scripts/install.sh | bash -s -- --alpha
```

Or with Go:

```sh
go install github.com/decampsrenan/spm@v0.3.0-alpha.1  # replace with the desired alpha tag
```

Alpha builds are published automatically on every push to `main`.

### From source

Requires [Go](https://go.dev/) 1.25+.

```sh
go install github.com/decampsrenan/spm@latest
```

## Usage

```sh
# Install dependencies (auto-detects npm/yarn/pnpm/bun/deno)
spm install

# Pass flags through to the underlying package manager
spm install --legacy-peer-deps
spm install --frozen-lockfile

# Add a package
spm add react

# Search and pick a package interactively
spm add

# Add a dev dependency
spm add vitest --save-dev

# Remove a package
spm remove react

# Run a script defined in package.json
spm dev
spm test
spm build

# Pick a script interactively from package.json
spm run

# Remove node_modules
spm clean

# Remove node_modules and the lock file
spm clean --lock

# Skip the confirmation prompt (useful in CI)
spm clean --yes

# Initialize a new project (interactive PM selection)
spm init

# Initialize with a specific package manager
spm init npm
spm init yarn

# Preview the init command without running it
spm init npm --dry-run

# Preview what would run without executing
spm dev --dry-run

# Install with background music 🎶
spm install --vibes

# Play a sound when the command finishes
spm install --notify

# Combine vibes and notification
spm install --vibes --notify

# Show raw package manager output (skip progress TUI)
spm install --raw

# Run a security audit on dependencies
spm audit

# Audit only production dependencies
spm audit --prod-only

# Output audit results as JSON
spm audit --json

# Only report high and critical vulnerabilities
spm audit --severity high

# Preview the audit command without running it
spm audit --dry-run

# Upgrade spm to the latest version
spm upgrade

# Upgrade to the latest alpha/pre-release
spm upgrade --alpha

# Reinstall even if already up to date
spm upgrade --force

# Preview upgrade without downloading
spm upgrade --dry-run

# Show version
spm --version
spm -v
```

### Command mapping

| spm command   | npm               | yarn            | pnpm            | bun             | deno              |
| ------------- | ----------------- | --------------- | --------------- | --------------- | ----------------- |
| `spm init`    | `npm init -y`     | `yarn init -y`  | `pnpm init`     | `bun init`      | `deno init`       |
| `spm install` | `npm install`     | `yarn install`  | `pnpm install`  | `bun install`   | `deno install`    |
| `spm add foo` | `npm install foo` | `yarn add foo`  | `pnpm add foo`  | `bun add foo`   | `deno add foo`    |
| `spm add`     | *(interactive search)* | *(interactive search)* | *(interactive search)* | *(interactive search)* | *(interactive search)* |
| `spm run`     | *(interactive)*   | *(interactive)* | *(interactive)* | *(interactive)* | *(interactive)*   |
| `spm remove foo` | `npm uninstall foo` | `yarn remove foo` | `pnpm remove foo` | `bun remove foo` | `deno remove foo` |
| `spm clean`   | Removes `node_modules` (and lock file with `--lock`) |                 |                 |                   |
| `spm audit`   | `npm audit --json`| `yarn audit --json` | `pnpm audit --json` |           |                   |
| `spm upgrade` | Self-updates spm via GitHub Releases |                 |                 |                   |
| `spm dev`     | `npm run dev`     | `yarn dev`      | `pnpm dev`      | `bun dev`       | `deno task dev`   |

## Contributing

Contributions are welcome! 🙌 Please read the [contributing guide](CONTRIBUTING.md) to get started.

## Support

Got a question or found something weird? Don't hesitate to [open an issue](https://github.com/DecampsRenan/spm/issues) — I'm happy to help! 💬

## License

Distributed under the [MIT](LICENSE) license.

## Acknowledgements

Big thanks to these awesome projects that make spm possible:

- [Cobra](https://github.com/spf13/cobra) — CLI framework
- [Charmbracelet](https://github.com/charmbracelet) — terminal styling, prompts, and spinner (lipgloss, huh, bubbletea, bubbles)
- [Beep](https://github.com/gopxl/beep) — audio playback for vibes & notifications
- [go-isatty](https://github.com/mattn/go-isatty) — TTY detection
