#!/bin/bash
set -euo pipefail

# Release script for supawho
#
# Usage:
#   ./release.sh patch    # 1.2.0 → 1.2.1  (bugfix)
#   ./release.sh minor    # 1.2.0 → 1.3.0  (new feature)
#   ./release.sh major    # 1.2.0 → 2.0.0  (breaking change)
#   ./release.sh          # default: patch
#
# What it does:
#   1. Bumps VERSION in the supawho script
#   2. Commits and tags the new version
#   3. Pushes commit + tag to GitHub
#   4. GitHub Action auto-updates the Homebrew formula

SCRIPT="supawho"
BUMP="${1:-patch}"

# Read current version
CURRENT=$(grep '^VERSION=' "$SCRIPT" | cut -d '"' -f 2)
IFS='.' read -r MAJOR MINOR PATCH <<< "$CURRENT"

case "$BUMP" in
  major) MAJOR=$((MAJOR + 1)); MINOR=0; PATCH=0 ;;
  minor) MINOR=$((MINOR + 1)); PATCH=0 ;;
  patch) PATCH=$((PATCH + 1)) ;;
  *)
    echo "Usage: ./release.sh [major|minor|patch]"
    exit 1
    ;;
esac

NEW="${MAJOR}.${MINOR}.${PATCH}"

echo -e "\n  📦 Releasing supawho"
echo -e "     ${CURRENT} → ${NEW} (${BUMP})\n"

# Update version in script
sed -i '' "s/VERSION=\"${CURRENT}\"/VERSION=\"${NEW}\"/" "$SCRIPT"

# Commit, tag, push
git add "$SCRIPT"
git commit -m "Bump version to ${NEW}"
git tag "v${NEW}"
git push
git push origin "v${NEW}"

echo -e "\n  ✅ v${NEW} released!"
echo -e "     GitHub Action will update the Homebrew formula automatically.\n"
