class GitPr < Formula
  desc "AI-powered PR title and description generator"
  homepage "https://github.com/evanyang0612/git-pr"
  version "1.0.0"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/evanyang0612/git-pr/releases/download/v#{version}/git-pr_darwin_arm64.tar.gz"
      sha256 "REPLACE_WITH_SHA256_AFTER_RELEASE"
    end
    on_intel do
      url "https://github.com/evanyang0612/git-pr/releases/download/v#{version}/git-pr_darwin_amd64.tar.gz"
      sha256 "REPLACE_WITH_SHA256_AFTER_RELEASE"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/evanyang0612/git-pr/releases/download/v#{version}/git-pr_linux_arm64.tar.gz"
      sha256 "REPLACE_WITH_SHA256_AFTER_RELEASE"
    end
    on_intel do
      url "https://github.com/evanyang0612/git-pr/releases/download/v#{version}/git-pr_linux_amd64.tar.gz"
      sha256 "REPLACE_WITH_SHA256_AFTER_RELEASE"
    end
  end

  depends_on "gh"

  def install
    bin.install "git-pr"
  end

  test do
    assert_match "AI-powered PR generator", shell_output("#{bin}/git-pr --help")
  end
end
