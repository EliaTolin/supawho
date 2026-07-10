#!/bin/bash
set -euo pipefail

REPO="EliaTolin/supawho"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
VERSION="${1:-latest}"

echo -e "\n  🔍 Installing supawho...\n"

# Detect OS and architecture
os="$(uname -s)"
arch="$(uname -m)"

case "$os" in
  Darwin) os="Darwin" ;;
  Linux)  os="Linux" ;;
  *)
    echo "  ❌ Unsupported OS: $os"
    echo "     On Windows, download the .zip from the releases page instead."
    exit 1
    ;;
esac

case "$arch" in
  x86_64|amd64) arch="x86_64" ;;
  arm64|aarch64) arch="arm64" ;;
  *)
    echo "  ❌ Unsupported architecture: $arch"
    exit 1
    ;;
esac

# Resolve the tag to install
if [ "$VERSION" = "latest" ]; then
  TAG=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" 2>/dev/null | grep '"tag_name"' | cut -d '"' -f 4 || true)
  if [ -z "$TAG" ]; then
    echo "  ❌ Could not resolve the latest release."
    exit 1
  fi
else
  TAG="v${VERSION#v}"
fi

VER="${TAG#v}"
ARCHIVE="supawho_${VER}_${os}_${arch}.tar.gz"
BASE_URL="https://github.com/${REPO}/releases/download/${TAG}"

TMPDIR="$(mktemp -d)"
trap 'rm -rf "$TMPDIR"' EXIT

echo "  → Downloading ${ARCHIVE}"
if ! curl -fsSL "${BASE_URL}/${ARCHIVE}" -o "${TMPDIR}/${ARCHIVE}"; then
  echo "  ❌ Download failed. Check that version '${VERSION}' has a build for ${os}/${arch}."
  exit 1
fi

# Verify checksum
if curl -fsSL "${BASE_URL}/checksums.txt" -o "${TMPDIR}/checksums.txt"; then
  echo "  → Verifying checksum"
  expected="$(grep " ${ARCHIVE}\$" "${TMPDIR}/checksums.txt" | awk '{print $1}')"
  if [ -n "$expected" ]; then
    if command -v shasum >/dev/null 2>&1; then
      actual="$(shasum -a 256 "${TMPDIR}/${ARCHIVE}" | awk '{print $1}')"
    else
      actual="$(sha256sum "${TMPDIR}/${ARCHIVE}" | awk '{print $1}')"
    fi
    if [ "$expected" != "$actual" ]; then
      echo "  ❌ Checksum mismatch — aborting."
      exit 1
    fi
  fi
fi

tar -xzf "${TMPDIR}/${ARCHIVE}" -C "$TMPDIR"

# Install
if [ -w "$INSTALL_DIR" ]; then
  mv "${TMPDIR}/supawho" "${INSTALL_DIR}/supawho"
  chmod +x "${INSTALL_DIR}/supawho"
else
  sudo mv "${TMPDIR}/supawho" "${INSTALL_DIR}/supawho"
  sudo chmod +x "${INSTALL_DIR}/supawho"
fi

echo -e "\n  ✅ supawho ${TAG} installed to ${INSTALL_DIR}/supawho\n"
echo -e "  Run 'supawho help' to get started.\n"
