## D1: Releases, install script and self-update
date: 2026-10-06
outcome: true

Decided by the user: GoReleaser builds binaries for macOS, Linux and Windows on tag push and publishes them as GitHub Releases with checksums; an install script downloads the right binary into ~/.local/bin; prep update replaces the binary from the latest release, or names the right command when prep was installed by Homebrew or go install; go install keeps working. The Claude Code integration ships as a plugin (20261005-184238) that checks the binary version. Implemented in 20261005-152624. Alternatives: go install only (needs a Go toolchain, no updates); Homebrew first (macOS and Linux only).
