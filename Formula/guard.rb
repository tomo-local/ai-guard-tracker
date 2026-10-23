class HabitTracker < Formula
  desc "ai-guard-tracker (command: guard) is a CLI tool that integrates with Claude Code's hook system to allow, deny, or require human confirmation for Bash commands run by AI coding agents based on repo-defined policy manifests."
  homepage "https://github.com/tomo-local/guard"
  version "0.0.1"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/tomo-local/guard/releases/download/v0.0.1/habit-tracker-darwin-arm64"
      sha256 "c45bffb6cb2406e00474d6f2288791b82f3ba947b009475f8728de03d276bb04"
    end
    on_intel do
      url "https://github.com/tomo-local/guard/releases/download/v0.0.1/habit-tracker-darwin-amd64"
      sha256 "cabc3f2c71186b9c6a3cb649c88b6ec26d726145e232ee949e097440cad048ba"
    end
  end

  on_linux do
    url "https://github.com/tomo-local/guard/releases/download/v0.0.2/habit-tracker-linux-amd64"
    sha256 "3458568ddbb80a0ff92147883217436c86768c70f1f19074d0f82fa48bfaffea"
  end

  def install
    binary_name = Dir["guard*"].first
    bin.install binary_name => "guard"
  end

  test do
    system "#{bin}/guard", "-h"
  end
end
