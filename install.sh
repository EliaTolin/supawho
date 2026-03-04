#!/bin/bash
set -euo pipefail

REPO="EliaTolin/supawho"
INSTALL_DIR="/usr/local/bin"
VERSION="${1:-latest}"

echo -e "\n  🔍 Installing supawho...\n"


# Resolve download URL
if [ "$VERSION" = "latest" ]; then
  TAG=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" 2>/dev/null | grep '"tag_name"' | cut -d '"' -f 4 || true)
  if [ -z "$TAG" ]; then
    URL="https://raw.githubusercontent.com/${REPO}/main/supawho"
  else
    URL="https://raw.githubusercontent.com/${REPO}/${TAG}/supawho"
  fi
else
  TAG="v${VERSION#v}"
  URL="https://raw.githubusercontent.com/${REPO}/${TAG}/supawho"
fi

# Download
TMPFILE=$(mktemp)
if ! curl -fsSL "$URL" -o "$TMPFILE"; then
  echo "  ❌ Download failed. Check that version '${VERSION}' exists."
  rm -f "$TMPFILE"
  exit 1
fi

# Install
if [ -w "$INSTALL_DIR" ]; then
  mv "$TMPFILE" "${INSTALL_DIR}/supawho"
  chmod +x "${INSTALL_DIR}/supawho"
else
  sudo mv "$TMPFILE" "${INSTALL_DIR}/supawho"
  sudo chmod +x "${INSTALL_DIR}/supawho"
fi

echo -e "  ✅ supawho installed to ${INSTALL_DIR}/supawho\n"
echo -e "  Run 'supawho help' to get started.\n"
