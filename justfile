# Install git hooks and dev tools
setup:
    #!/usr/bin/env bash
    set -euo pipefail
    go install golang.org/x/tools/cmd/goimports@latest
    go install github.com/mitranim/gow@latest
    HOOKS_DIR="$(git rev-parse --git-path hooks)"
    mkdir -p "$HOOKS_DIR"
    cp scripts/pre-commit "$HOOKS_DIR/pre-commit"
    chmod +x "$HOOKS_DIR/pre-commit"
    echo "Git hooks installed."

# Format all Go files
fmt:
    goimports -w .

# Run tests
test:
    go test ./... -v -race

# Build binary
build:
    go build -o spm .

# Watch and rebuild dev binary
dev:
    gow -e=go,mod build -ldflags "-X main.version=dev" -o "$(go env GOPATH)/bin/spm-dev" .
