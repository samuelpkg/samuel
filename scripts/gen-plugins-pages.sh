#!/usr/bin/env bash
# gen-plugins-pages.sh — generate one docs/plugins/<name>.md per registry entry
#
# Usage: scripts/gen-plugins-pages.sh [--source <file-or-url>]
#
# Fetches samuelpkg/samuel-registry's index.toml (or a local source via --source),
# parses [[plugins]] blocks, writes one short page per plugin pointing to its repo.
# Idempotent: stable output across reruns.
#
# Known limitation (v2.0): does NOT update mkdocs.yml's `Plugins:` nav section.
# Pages exist on disk after this script runs but must be manually added to the
# nav until v2.1 ships the nav-injector.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
cd "${REPO_ROOT}"

DEFAULT_SOURCE="https://raw.githubusercontent.com/samuelpkg/samuel-registry/main/index.toml"
SOURCE="${DEFAULT_SOURCE}"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --source) SOURCE="$2"; shift 2 ;;
    -h|--help) sed -n '2,12p' "$0"; exit 0 ;;
    *) echo "unknown flag: $1" >&2; exit 2 ;;
  esac
done

OUT_DIR="docs/plugins"
INDEX_PAGE="${OUT_DIR}/index.md"
TMP_TOML="$(mktemp)"
trap 'rm -f "${TMP_TOML}"' EXIT

# Fetch or copy the source TOML.
if [[ "${SOURCE}" == http*://* ]]; then
  if command -v curl >/dev/null 2>&1; then
    curl -sSL --fail "${SOURCE}" -o "${TMP_TOML}"
  elif command -v wget >/dev/null 2>&1; then
    wget -q "${SOURCE}" -O "${TMP_TOML}"
  else
    echo "error: neither curl nor wget available" >&2
    exit 1
  fi
elif [[ "${SOURCE}" == file://* ]]; then
  cp "${SOURCE#file://}" "${TMP_TOML}"
else
  cp "${SOURCE}" "${TMP_TOML}"
fi

mkdir -p "${OUT_DIR}"

# Parse [[plugins]] blocks. Emit TSV: name<TAB>kind<TAB>version<TAB>repo<TAB>description<TAB>tags-csv
ENTRIES="$(
  awk '
    function strip_quotes(s) {
      sub(/^[[:space:]]*"/, "", s)
      sub(/"[[:space:]]*$/, "", s)
      return s
    }
    function strip_array(s,   inner, n, i, parts, out) {
      sub(/^[[:space:]]*\[/, "", s)
      sub(/\][[:space:]]*$/, "", s)
      n = split(s, parts, ",")
      out = ""
      for (i = 1; i <= n; i++) {
        gsub(/^[[:space:]]+|[[:space:]]+$/, "", parts[i])
        gsub(/"/, "", parts[i])
        if (parts[i] == "") continue
        out = (out == "") ? parts[i] : out "," parts[i]
      }
      return out
    }
    function emit() {
      if (name == "") return
      printf "%s\t%s\t%s\t%s\t%s\t%s\n", name, kind, version, repo, description, tags
      name = ""; kind = ""; version = ""; repo = ""; description = ""; tags = ""
    }
    /^\[\[plugins\]\]/ { emit(); next }
    /^[[:space:]]*name[[:space:]]*=/ {
      sub(/^[[:space:]]*name[[:space:]]*=[[:space:]]*/, "", $0); name = strip_quotes($0); next
    }
    /^[[:space:]]*kind[[:space:]]*=/ {
      sub(/^[[:space:]]*kind[[:space:]]*=[[:space:]]*/, "", $0); kind = strip_quotes($0); next
    }
    /^[[:space:]]*version[[:space:]]*=/ {
      sub(/^[[:space:]]*version[[:space:]]*=[[:space:]]*/, "", $0); version = strip_quotes($0); next
    }
    /^[[:space:]]*latest[[:space:]]*=/ {
      sub(/^[[:space:]]*latest[[:space:]]*=[[:space:]]*/, "", $0); version = strip_quotes($0); next
    }
    /^[[:space:]]*repo[[:space:]]*=/ {
      sub(/^[[:space:]]*repo[[:space:]]*=[[:space:]]*/, "", $0); repo = strip_quotes($0); next
    }
    /^[[:space:]]*source[[:space:]]*=/ {
      sub(/^[[:space:]]*source[[:space:]]*=[[:space:]]*/, "", $0); repo = strip_quotes($0); next
    }
    /^[[:space:]]*description[[:space:]]*=/ {
      sub(/^[[:space:]]*description[[:space:]]*=[[:space:]]*/, "", $0); description = strip_quotes($0); next
    }
    /^[[:space:]]*tags[[:space:]]*=/ {
      sub(/^[[:space:]]*tags[[:space:]]*=[[:space:]]*/, "", $0); tags = strip_array($0); next
    }
    END { emit() }
  ' "${TMP_TOML}"
)"

if [[ -z "${ENTRIES}" ]]; then
  echo "error: no [[plugins]] entries found in source TOML" >&2
  exit 1
fi

COUNT=0
CHANGED=0
WROTE=()

while IFS=$'\t' read -r NAME KIND VERSION REPO DESCRIPTION TAGS; do
  [[ -z "${NAME}" ]] && continue
  COUNT=$((COUNT + 1))
  PAGE="${OUT_DIR}/${NAME}.md"
  TMP_PAGE="$(mktemp)"

  REPO_LINK="${REPO:-https://github.com/samuelpkg/${NAME}}"
  TAGS_DISPLAY="${TAGS:-—}"
  DESC_DISPLAY="${DESCRIPTION:-No description provided.}"

  {
    echo "# ${NAME}"
    echo
    echo "- **Kind**: ${KIND:-skill}"
    echo "- **Latest**: ${VERSION:-unknown}"
    echo "- **Repo**: <${REPO_LINK}>"
    echo "- **Tags**: ${TAGS_DISPLAY//,/, }"
    echo
    echo "${DESC_DISPLAY}"
    echo
    echo "## Install"
    echo
    echo '```bash'
    echo "samuel install ${NAME}"
    echo '```'
    echo
    echo "## See also"
    echo
    echo "- [Registry entry](https://github.com/samuelpkg/samuel-registry/blob/main/index.toml)"
    echo "- [Plugin authors](../plugin-authors/index.md)"
    echo
    echo "<!-- gen-plugins-pages.sh: re-run to refresh from registry. Hand-edits below this marker survive regeneration only if you preserve the marker. -->"
  } > "${TMP_PAGE}"

  if [[ -f "${PAGE}" ]] && cmp -s "${TMP_PAGE}" "${PAGE}"; then
    rm -f "${TMP_PAGE}"
  else
    mv "${TMP_PAGE}" "${PAGE}"
    CHANGED=$((CHANGED + 1))
    WROTE+=("${PAGE}")
  fi
done <<< "${ENTRIES}"

echo "plugins pages: ${COUNT} entries processed, ${CHANGED} written/updated"

# Preserve the human-curated index.md if it exists — never overwrite it.
if [[ ! -f "${INDEX_PAGE}" ]]; then
  echo "warn: ${INDEX_PAGE} missing; create it manually (see Plugins:Index in mkdocs.yml)" >&2
fi
