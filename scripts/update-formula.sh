#!/usr/bin/env bash
set -euo pipefail

# Regenerates Formula/habit-tracker.rb for the given tag using the
# release assets already uploaded to GitHub Releases.
#
# Usage: scripts/update-formula.sh v1.2.3

TAG="${1:?usage: update-formula.sh <tag>}"
VERSION="${TAG#v}"
APP="guard"
REPO="tomo-local/${APP}"

WORKDIR="$(mktemp -d)"
trap 'rm -rf "$WORKDIR"' EXIT

gh release download "$TAG" --repo "$REPO" --dir "$WORKDIR" \
  --pattern "$APP-darwin-amd64" \
  --pattern "$APP-darwin-arm64" \
  --pattern "$APP-linux-amd64" \
  --clobber

sha_darwin_amd64="$(shasum -a 256 "$WORKDIR/$APP-darwin-amd64" | cut -d' ' -f1)"
sha_darwin_arm64="$(shasum -a 256 "$WORKDIR/$APP-darwin-arm64" | cut -d' ' -f1)"
sha_linux_amd64="$(shasum -a 256 "$WORKDIR/$APP-linux-amd64" | cut -d' ' -f1)"

cat > Formula/$APP.rb <<EOF
class App < Formula
  desc "ai-guard-tracker (command: guard) is a CLI tool that integrates with Claude Code's hook system to allow, deny, or require human confirmation for Bash commands run by AI coding agents based on repo-defined policy manifests."
  homepage "https://github.com/${REPO}"
  version "${VERSION}"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/${REPO}/releases/download/${TAG}/${APP}-darwin-arm64"
      sha256 "${sha_darwin_arm64}"
    end
    on_intel do
      url "https://github.com/${REPO}/releases/download/${TAG}/${APP}-darwin-amd64"
      sha256 "${sha_darwin_amd64}"
    end
  end

  on_linux do
    url "https://github.com/${REPO}/releases/download/${TAG}/${APP}-linux-amd64"
    sha256 "${sha_linux_amd64}"
  end

  def install
    binary_name = Dir["${APP}-*"].first
    bin.install binary_name => "${APP}"
  end

  test do
    system "#{bin}/${APP}", "-h"
  end
end
EOF

echo "Formula/${APP}.rb updated for ${TAG}"
