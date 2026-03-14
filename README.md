<!-- NOTE FOR AGENTS: Keep this README up to date when adding new commands,
     flags, or changing behavior. Update the usage examples and feature list. -->

<h1 align="center">
  📦 spm
</h1>

<div align="center">
  <strong>Smart Package Manager</strong> — One CLI to rule them all. Stop worrying about npm vs yarn vs pnpm, just run <code>spm</code> and let it figure out the rest.
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

- 🔍 **Auto-detection** via lock files (`package-lock.json`, `yarn.lock`, `pnpm-lock.yaml`)
- 📂 **Directory walk-up** — works from any subdirectory in your project
- 🔄 **Command translation** — maps commands to the correct syntax for each package manager
- 💬 **Interactive prompt** when multiple lock files are detected
- 👀 **Dry-run mode** to preview commands without executing them
- 🎵 **Vibes mode** — play background music while installing dependencies (`--vibes`)
- 🔔 **Notification sounds** — get notified when the command finishes (`--notify`)

### Built With

- [Go](https://go.dev/)
- [Cobra](https://github.com/spf13/cobra)
- [Beep](https://github.com/gopxl/beep)

## Getting Started

### Install script (macOS / Linux)

The fastest way to get started:

```sh
curl -fsSL https://raw.githubusercontent.com/decampsrenan/spm/main/scripts/install.sh | bash
```

Want to install to a custom directory? No problem:

```sh
INSTALL_DIR=~/.local/bin curl -fsSL https://raw.githubusercontent.com/decampsrenan/spm/main/scripts/install.sh | bash
```

### From source

Requires [Go](https://go.dev/) 1.25+.

```sh
go install github.com/decampsrenan/spm@latest
```

## Usage

```sh
# Install dependencies (auto-detects npm/yarn/pnpm)
spm install

# Add a package
spm add react

# Add a dev dependency
spm add vitest --save-dev

# Run a script defined in package.json
spm dev
spm test
spm build

# Preview what would run without executing
spm dev --dry-run

# Install with background music 🎶
spm install --vibes

# Play a sound when the command finishes
spm install --notify

# Combine vibes and notification
spm install --vibes --notify

# Show version
spm --version
spm -v
```

### Command mapping

| spm command   | npm               | yarn            | pnpm            |
| ------------- | ----------------- | --------------- | --------------- |
| `spm install` | `npm install`     | `yarn install`  | `pnpm install`  |
| `spm add foo` | `npm install foo` | `yarn add foo`  | `pnpm add foo`  |
| `spm dev`     | `npm run dev`     | `yarn dev`      | `pnpm dev`      |

## Contributing

Contributions are welcome! 🙌 Please read the [contributing guide](CONTRIBUTING.md) to get started.

## Support

Got a question or found something weird? Don't hesitate to [open an issue](https://github.com/DecampsRenan/spm/issues) — I'm happy to help! 💬

## License

Distributed under the [MIT](LICENSE) license.

## Acknowledgements

Big thanks to these awesome projects that make spm possible:

- [Cobra](https://github.com/spf13/cobra) — CLI framework
- [Beep](https://github.com/gopxl/beep) — audio playback for vibes & notifications
- [Survey](https://github.com/AlecAivazis/survey) — interactive terminal prompts
- [go-isatty](https://github.com/mattn/go-isatty) — TTY detection
