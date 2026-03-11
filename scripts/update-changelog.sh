#!/usr/bin/env bash
set -euo pipefail

# Updates CHANGELOG.md by moving [Unreleased] entries into a new versioned section.
# Usage: ./scripts/update-changelog.sh v0.2.0

VERSION="$1"
VERSION_NUM="${VERSION#v}"
DATE=$(date +%Y-%m-%d)
FILE="CHANGELOG.md"

# Check that [Unreleased] section has entries
UNRELEASED=$(sed -n '/^## \[Unreleased\]/,/^## \[/p' "$FILE" | tail -n +2 | sed '$d')
if ! echo "$UNRELEASED" | grep -q '^ *- '; then
  echo "No unreleased entries found, skipping CHANGELOG update."
  exit 0
fi

# Find previous version tag from CHANGELOG headings
PREV_VERSION=$(grep -oE '## \[[0-9]+\.[0-9]+\.[0-9]+\]' "$FILE" | head -1 | grep -oE '[0-9]+\.[0-9]+\.[0-9]+')

# Replace [Unreleased] section: insert empty [Unreleased] + new version heading
awk -v ver="$VERSION_NUM" -v date="$DATE" '
BEGIN { in_unreleased = 0; printed_new = 0 }
/^## \[Unreleased\]/ {
  print "## [Unreleased]"
  print ""
  in_unreleased = 1
  next
}
in_unreleased && /^## \[/ {
  printf "## [%s] - %s\n", ver, date
  print ""
  in_unreleased = 0
  printed_new = 1
}
in_unreleased {
  # Buffer unreleased content, skip leading blank lines
  if (buf == "" && $0 == "") next
  buf = buf $0 "\n"
  next
}
printed_new == 1 {
  printf "%s", buf
  buf = ""
  printed_new = 0
}
{ print }
' "$FILE" > "${FILE}.tmp" && mv "${FILE}.tmp" "$FILE"

# Update comparison links at bottom
# [Unreleased]: .../compare/vOLD...HEAD -> .../compare/vNEW...HEAD
sed -i.bak "s|\[Unreleased\]: \(.*\)/compare/v${PREV_VERSION}\.\.\.HEAD|[Unreleased]: \1/compare/${VERSION}...HEAD|" "$FILE"

# Insert new version link before previous version link
sed -i.bak "/^\[${PREV_VERSION}\]:/i\\
[${VERSION_NUM}]: https://github.com/DecampsRenan/spm/compare/v${PREV_VERSION}...${VERSION}
" "$FILE"

# Clean up backup files from sed -i
rm -f "${FILE}.bak"

echo "CHANGELOG.md updated for ${VERSION}"
