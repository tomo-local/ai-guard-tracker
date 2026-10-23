class Guard < Formula
  desc "ai-guard-tracker (command: guard) is a CLI tool that integrates with Claude Code's hook system to allow, deny, or require human confirmation for Bash commands run by AI coding agents based on repo-defined policy manifests."
  homepage "https://github.com/tomo-local/guard"
  version "0.0.3"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/tomo-local/guard/releases/download/v0.0.3/guard-darwin-arm64"
      sha256 "848d45648bd7a80cdf554848e3fccc9c3066c91578a5be37672527dec5470144"
    end
    on_intel do
      url "https://github.com/tomo-local/guard/releases/download/v0.0.3/guard-darwin-amd64"
      sha256 "2ef9109b6fb282d1e74cfd85f3b141b7e8e0ca8a61fb8c8e86fb3d5334b538ce"
    end
  end

  on_linux do
    url "https://github.com/tomo-local/guard/releases/download/v0.0.3/guard-linux-amd64"
    sha256 "6a7c828659e0edf208807f3aa53a68aeae5b512a9d03d36b66c6b7c90dc78f72"
  end

  def install
    binary_name = Dir["guard-*"].first
    bin.install binary_name => "guard"
  end

  test do
    system "#{bin}/guard", "-h"
  end
end
