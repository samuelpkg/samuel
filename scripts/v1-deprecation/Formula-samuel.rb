# This file is REFERENCE ONLY — goreleaser generates the real formula.
#
# The brews block in .goreleaser.yaml writes Formula/samuel.rb directly
# into github.com/ar4mirez/homebrew-tap on each tag push, with the
# per-architecture URLs and SHA-256 sums filled in from the actual
# release artifacts. Do not edit a formula by hand and commit it to the
# tap — goreleaser will overwrite it on the next release.
#
# The shape goreleaser produces (rendered from the install/test blocks
# in .goreleaser.yaml's brews section) looks roughly like this:
#
#   class Samuel < Formula
#     desc "Rails for AI coding assistants"
#     homepage "https://samuelpkg.github.io/samuel/"
#     version "2.0.0"
#     license "MIT"
#
#     on_macos do
#       on_arm do
#         url "https://github.com/samuelpkg/samuel/releases/download/v2.0.0/samuel_2.0.0_darwin_arm64.tar.gz"
#         sha256 "..."
#       end
#       on_intel do
#         url "https://github.com/samuelpkg/samuel/releases/download/v2.0.0/samuel_2.0.0_darwin_amd64.tar.gz"
#         sha256 "..."
#       end
#     end
#
#     on_linux do
#       on_arm do
#         url "https://github.com/samuelpkg/samuel/releases/download/v2.0.0/samuel_2.0.0_linux_arm64.tar.gz"
#         sha256 "..."
#       end
#       on_intel do
#         url "https://github.com/samuelpkg/samuel/releases/download/v2.0.0/samuel_2.0.0_linux_amd64.tar.gz"
#         sha256 "..."
#       end
#     end
#
#     def install
#       bin.install "samuel"
#       generate_completions_from_executable(bin/"samuel", "completion")
#     end
#
#     test do
#       system "#{bin}/samuel", "version"
#     end
#   end
#
# Token required at release time:
#   HOMEBREW_TAP_GITHUB_TOKEN  (fine-grained PAT, scoped to
#                               ar4mirez/homebrew-tap, Contents: r/w).
#
# Verify after publication:
#   brew update
#   brew install ar4mirez/tap/samuel
#   samuel version    # prints v2.0.0
