#!/usr/bin/env bash
# gen-rfd-index.sh — regenerate docs/rfd/index.md's RFD table from rfd-index.toml
#
# Usage: scripts/gen-rfd-index.sh
#
# Idempotent: re-running with no TOML changes produces zero diff.
# Dependency-light: uses awk + sed only.

set -euo pipefail

# Resolve repo root from this script's location.
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
cd "${REPO_ROOT}"

TOML="rfd-index.toml"
INDEX="docs/rfd/index.md"
START_MARKER="<!-- RFD_INDEX_START -->"
END_MARKER="<!-- RFD_INDEX_END -->"

if [[ ! -f "${TOML}" ]]; then
  echo "error: ${TOML} not found at repo root" >&2
  exit 1
fi

if [[ ! -f "${INDEX}" ]]; then
  echo "error: ${INDEX} not found" >&2
  exit 1
fi

# Parse [[rfds]] blocks from rfd-index.toml into a TSV stream:
#   number<TAB>title<TAB>state<TAB>labels-csv
# awk walks line by line, accumulating per-block fields,
# emitting on the next [[rfds]] header (and once at EOF).
TABLE_BODY="$(
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
        out = (out == "") ? parts[i] : out ", " parts[i]
      }
      return out
    }
    function emit() {
      if (number == "") return
      printf "| [%s](%s.md) | %s | %s | %s |\n", number, number, title, state, labels
      number = ""; title = ""; state = ""; labels = ""
    }
    /^\[\[rfds\]\]/ { emit(); next }
    /^[[:space:]]*number[[:space:]]*=/ {
      sub(/^[[:space:]]*number[[:space:]]*=[[:space:]]*/, "", $0)
      number = strip_quotes($0); next
    }
    /^[[:space:]]*title[[:space:]]*=/ {
      sub(/^[[:space:]]*title[[:space:]]*=[[:space:]]*/, "", $0)
      title = strip_quotes($0); next
    }
    /^[[:space:]]*state[[:space:]]*=/ {
      sub(/^[[:space:]]*state[[:space:]]*=[[:space:]]*/, "", $0)
      state = strip_quotes($0); next
    }
    /^[[:space:]]*labels[[:space:]]*=/ {
      sub(/^[[:space:]]*labels[[:space:]]*=[[:space:]]*/, "", $0)
      labels = strip_array($0); next
    }
    END { emit() }
  ' "${TOML}"
)"

if [[ -z "${TABLE_BODY}" ]]; then
  echo "error: parsed zero RFD entries from ${TOML}" >&2
  exit 1
fi

# Build the replacement block.
REPLACEMENT_FILE="$(mktemp)"
trap 'rm -f "${REPLACEMENT_FILE}"' EXIT
{
  echo "${START_MARKER}"
  echo
  echo "| # | Title | State | Labels |"
  echo "| --- | --- | --- | --- |"
  echo "${TABLE_BODY}"
  echo
  echo "${END_MARKER}"
} > "${REPLACEMENT_FILE}"

# Splice the block between the markers in INDEX.
OUT_FILE="$(mktemp)"
trap 'rm -f "${REPLACEMENT_FILE}" "${OUT_FILE}"' EXIT

awk -v repl_file="${REPLACEMENT_FILE}" \
    -v start_marker="${START_MARKER}" \
    -v end_marker="${END_MARKER}" '
  BEGIN {
    while ((getline line < repl_file) > 0) {
      repl = (repl == "") ? line : repl "\n" line
    }
    close(repl_file)
    inside = 0
  }
  {
    if (index($0, start_marker) > 0) {
      print repl
      inside = 1
      next
    }
    if (inside && index($0, end_marker) > 0) {
      inside = 0
      next
    }
    if (!inside) print
  }
' "${INDEX}" > "${OUT_FILE}"

# Only overwrite if content changed (stable diffs).
if cmp -s "${OUT_FILE}" "${INDEX}"; then
  echo "rfd index: no changes"
else
  mv "${OUT_FILE}" "${INDEX}"
  echo "rfd index: updated ${INDEX}"
fi
