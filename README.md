# spm

<!-- NOTE FOR AGENTS: Keep this README up to date when adding new commands,
     flags, or changing behavior. Update the usage examples and feature list. -->

**Smart Package Manager** — a unified CLI that auto-detects your project's JavaScript package manager (npm, yarn, or pnpm) and proxies commands transparently.

Stop thinking about which package manager a project uses. Just run `spm install`, `spm add react`, or `spm dev` and let it figure out the rest.

## Features

- **Auto-detection** via lock files (`package-lock.json`, `yarn.lock`, `pnpm-lock.yaml`)
- **Directory walk-up** — works from any subdirectory in your project
- **Command translation** — maps commands to the correct syntax for each package manager
- **Interactive prompt** when multiple lock files are detected
- **Dry-run mode** to preview commands without executing them

## Installation

### Install script (macOS / Linux)

```sh
curl -fsSL https://raw.githubusercontent.com/decampsrenan/spm/main/scripts/install.sh | bash
```

To install to a custom directory:

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
```

### Command mapping

| spm command   | npm                   | yarn            | pnpm            |
| ------------- | --------------------- | --------------- | --------------- |
| `spm install` | `npm install`         | `yarn install`  | `pnpm install`  |
| `spm add foo` | `npm install foo`     | `yarn add foo`  | `pnpm add foo`  |
| `spm dev`     | `npm run dev`         | `yarn dev`      | `pnpm dev`      |

## Support

Found a bug or have a question? [Open an issue](https://github.com/decampsrenan/spm/issues).

## Contributing

Requires [Go](https://go.dev/) 1.25+ and [just](https://github.com/casey/just).

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
```

## License

[MIT](LICENSE)
