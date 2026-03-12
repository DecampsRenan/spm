# Install git hooks and dev tools
setup:
    #!/usr/bin/env bash
    set -euo pipefail
    go install golang.org/x/tools/cmd/goimports@latest
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

# Build and install as spm-dev in GOPATH/bin
dev:
    go build -o "$(go env GOPATH)/bin/spm-dev" .
