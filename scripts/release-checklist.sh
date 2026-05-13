#!/usr/bin/env bash
# Samuel v2 release checklist runner.
#
# Runs every preflight gate before a tag is cut. Does NOT mutate
# state — no tag, no push, no force-push. Reports green / red and
# exits non-zero on any failure.
#
# Usage:
#   scripts/release-checklist.sh                 # full preflight
#   scripts/release-checklist.sh --candidate rc.2  # rc.2 mode
#   scripts/release-checklist.sh --candidate rc.3  # rc.3 mode
#   scripts/release-checklist.sh --candidate ga    # v2.0.0 final
#
# After this exits 0, the operator runs (manually):
#   git tag -s v2.0.0-rc.N -m "..."
#   git push origin v2.0.0-rc.N
# and goreleaser publishes the signed artifacts.

set -euo pipefail

CANDIDATE="${1:-}"
if [ "$CANDIDATE" = "--candidate" ]; then
    CANDIDATE="${2:-}"
fi

cd "$(git rev-parse --show-toplevel)"

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

ok()    { printf "${GREEN}OK${NC}    %s\n" "$1"; }
fail()  { printf "${RED}FAIL${NC}  %s\n" "$1"; FAILURES=$((FAILURES + 1)); }
warn()  { printf "${YELLOW}WARN${NC}  %s\n" "$1"; }
step()  { printf "\n${YELLOW}==>${NC} %s\n" "$1"; }

FAILURES=0

step "Working tree is clean"
if [ -z "$(git status --porcelain)" ]; then
    ok "no uncommitted changes"
else
    fail "uncommitted changes present"
    git status --short
fi

step "Build succeeds"
if go build ./... 2>/dev/null; then
    ok "go build ./..."
else
    fail "go build ./... failed"
fi

step "Tests pass"
if go test -race ./... > /tmp/samuel-release-test.log 2>&1; then
    ok "go test -race ./..."
else
    fail "tests failed (see /tmp/samuel-release-test.log)"
fi

step "Lint clean (if golangci-lint installed)"
if command -v golangci-lint >/dev/null 2>&1; then
    if golangci-lint run ./... > /tmp/samuel-release-lint.log 2>&1; then
        ok "golangci-lint run ./..."
    else
        fail "golangci-lint run failed (see /tmp/samuel-release-lint.log)"
    fi
else
    warn "golangci-lint not installed — CI will run it"
fi

step "AGENTS.md template ≤ 150 lines (source)"
LINES=$(wc -l < template/AGENTS.md.tmpl)
if [ "$LINES" -le 150 ]; then
    ok "template/AGENTS.md.tmpl is $LINES lines"
else
    fail "template/AGENTS.md.tmpl is $LINES lines (budget: 150)"
fi

step "Agnostic-by-design grep clean"
PATTERNS='CLAUDE\.md|\.claude/|\.cursor/|\.codex/|Cursor|Codex CLI'
LEAKS=$(grep -RInE "$PATTERNS" internal/ cmd/ 2>/dev/null \
    | grep -v -E '_test\.go:' \
    | grep -v 'agnostic-allow' \
    || true)
if [ -z "$LEAKS" ]; then
    ok "no tool-specific leaks in internal/ or cmd/"
else
    fail "tool-specific references found:"
    echo "$LEAKS"
fi

step "RFD index up to date"
if bash scripts/gen-rfd-index.sh > /tmp/samuel-rfd-gen.log 2>&1; then
    if [ -z "$(git status --porcelain docs/rfd/index.md)" ]; then
        ok "docs/rfd/index.md matches rfd-index.toml"
    else
        fail "docs/rfd/index.md is stale (run scripts/gen-rfd-index.sh and commit)"
    fi
else
    fail "gen-rfd-index.sh failed (see /tmp/samuel-rfd-gen.log)"
fi

step "CHANGELOG has an entry for this candidate"
if [ -n "$CANDIDATE" ]; then
    case "$CANDIDATE" in
        rc.2|rc.3) PATTERN="v2.0.0-${CANDIDATE}" ;;
        ga)         PATTERN="v2.0.0" ;;
        *)          PATTERN="v2.0.0" ;;
    esac
    if grep -q "^## \[${PATTERN}\]" CHANGELOG.md; then
        ok "CHANGELOG has ## [${PATTERN}] entry"
    else
        fail "CHANGELOG.md missing entry for ${PATTERN}"
    fi
else
    if grep -q "^## \[v2.0.0\]" CHANGELOG.md; then
        ok "CHANGELOG has the v2.0.0 entry"
    else
        fail "CHANGELOG.md missing the v2.0.0 entry"
    fi
fi

step "goreleaser config validates"
if command -v goreleaser >/dev/null 2>&1; then
    if goreleaser check --config .goreleaser.yaml > /tmp/samuel-goreleaser.log 2>&1; then
        ok "goreleaser check"
    else
        fail "goreleaser check failed (see /tmp/samuel-goreleaser.log)"
    fi
else
    warn "goreleaser not installed locally — CI will run it on tag"
fi

step "Required artifacts exist"
for f in CHANGELOG.md README.md LICENSE install.sh .goreleaser.yaml \
         mkdocs.yml docs/index.md docs/rfd/index.md \
         scripts/v1-deprecation/README-v1-final.md \
         scripts/v1-deprecation/Formula-samuel.rb \
         template/AGENTS.md.tmpl rfd-index.toml; do
    if [ -f "$f" ]; then
        ok "$f"
    else
        fail "missing: $f"
    fi
done

step "Summary"
if [ "$FAILURES" -eq 0 ]; then
    printf "${GREEN}ALL GREEN${NC}  — preflight passed for ${CANDIDATE:-v2.0.0}\n"
    cat <<EOF

Next steps (manual):
  1. Tag and push:
       git tag -s v2.0.0${CANDIDATE:+-${CANDIDATE}} -m "v2.0.0${CANDIDATE:+-${CANDIDATE}}"
       git push origin v2.0.0${CANDIDATE:+-${CANDIDATE}}
  2. CI / goreleaser publishes signed artifacts.
  3. (Final release only) Tag v1-final at the last v1 commit before
     force-pushing main to v2; update the homebrew tap formula.
EOF
    exit 0
else
    printf "${RED}${FAILURES} FAILURE(S)${NC} — fix before tagging\n"
    exit 1
fi
