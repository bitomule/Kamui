class Kamui < Formula
  desc "🎯 Advanced session manager for Claude Code with automatic status line integration"
  homepage "https://github.com/bitomule/kamui"
  url "https://github.com/bitomule/kamui.git",
      tag:      "v1.0.0",
      revision: "d54e212a1b8c8f3e9d4f5a6b7c8d9e0f1a2b3c4d"
  license "MIT"
  head "https://github.com/bitomule/kamui.git", branch: "main"

  depends_on "go" => :build
  depends_on "node" => :recommended

  def install
    system "go", "build", "-ldflags", "-s -w", "-o", "kam", "cmd/kam/main.go"
    bin.install "kam"

    # Generate shell completions
    generate_completions_from_executable(bin/"kam", "completion")
  end

  def caveats
    <<~EOS
      🎯 Kamui is ready! 

      Quick start:
        kam MyProject          # Create/resume a session
        kam                   # Interactive session picker
        kam setup            # Configure Claude Code integration

      Requirements:
        • Claude Code CLI must be installed
        • Download from: https://claude.ai/code

      The status line will be configured automatically on first use.
    EOS
  end

  test do
    system "#{bin}/kam", "--version"
    system "#{bin}/kam", "--help"
  end
end