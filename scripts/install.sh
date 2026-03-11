#!/usr/bin/env bash
set -euo pipefail

REPO="decampsrenan/spm"

# Detect OS
case "$(uname -s)" in
  Linux*)  OS=linux;;
  Darwin*) OS=darwin;;
  *)       echo "Unsupported OS: $(uname -s)" && exit 1;;
esac

# Detect arch
case "$(uname -m)" in
  x86_64)  ARCH=amd64;;
  aarch64|arm64) ARCH=arm64;;
  *)       echo "Unsupported architecture: $(uname -m)" && exit 1;;
esac

# Get latest version
VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
if [ -z "$VERSION" ]; then
  echo "Failed to fetch latest version"
  exit 1
fi

FILENAME="spm_${VERSION#v}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${VERSION}/${FILENAME}"

INSTALL_DIR_SET="${INSTALL_DIR:+1}"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

echo "Downloading spm ${VERSION} for ${OS}/${ARCH}..."
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

curl -fsSL "$URL" -o "${TMP}/${FILENAME}"
tar -xzf "${TMP}/${FILENAME}" -C "$TMP"

install_spm() {
  install -d "$1" && install "${TMP}/spm" "$1/spm"
}

echo "Installing to ${INSTALL_DIR}/spm..."
if ! install_spm "$INSTALL_DIR" 2>/dev/null; then
  if [ -z "${INSTALL_DIR_SET:-}" ]; then
    INSTALL_DIR="$HOME/.local/bin"
    echo "No permission for /usr/local/bin, installing to ${INSTALL_DIR} instead..."
    install_spm "$INSTALL_DIR"
  else
    echo "Permission denied. Try with sudo or set INSTALL_DIR to a writable location."
    exit 1
  fi
fi

echo "spm ${VERSION} installed successfully!"

# Warn if INSTALL_DIR is not in PATH
case ":$PATH:" in
  *":${INSTALL_DIR}:"*) ;;
  *) echo "WARNING: ${INSTALL_DIR} is not in your PATH. Add it with:"
     echo "  export PATH=\"${INSTALL_DIR}:\$PATH\"";;
esac
