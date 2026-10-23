#!/usr/bin/env bash
set -euo pipefail

# Generates release notes for TAG by grouping commit messages since the
# previous tag by Conventional Commits type (feat/fix/docs/refactor/chore).
#
# Usage: scripts/generate-release-notes.sh v1.2.3

TAG="${1:?usage: generate-release-notes.sh <tag>}"
APP="guard"
REPO="tomo-local/${APP}"

PREV_TAG="$(git tag --sort=-v:refname | grep -Fxv "$TAG" | head -n1 || true)"
RANGE="${PREV_TAG:+${PREV_TAG}..}${TAG}"

print_section() {
  local heading="$1" pattern="$2"
  local commits
  commits="$(git log "$RANGE" --pretty=format:'%s' | grep -E "$pattern" || true)"
  if [ -n "$commits" ]; then
    echo "## ${heading}"
    echo "$commits" | sed -E 's/^[a-z]+(\([^)]*\))?: */- /'
    echo
  fi
}

print_section "Features" '^feat(\([^)]*\))?:'
print_section "Fixes" '^fix(\([^)]*\))?:'
print_section "Docs" '^docs(\([^)]*\))?:'
print_section "Refactor" '^refactor(\([^)]*\))?:'
print_section "Chores" '^chore(\([^)]*\))?:'

other="$(git log "$RANGE" --pretty=format:'%s' | grep -Ev '^(feat|fix|docs|refactor|chore)(\([^)]*\))?:' || true)"
if [ -n "$other" ]; then
  echo "## Other"
  echo "$other" | sed -E 's/^/- /'
  echo
fi

if [ -n "$PREV_TAG" ]; then
  echo "**Full Changelog**: https://github.com/${REPO}/compare/${PREV_TAG}...${TAG}"
fi
