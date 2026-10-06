## D1: Releases, install script and self-update
date: 2026-10-06
outcome: true

Decided by the user: GoReleaser builds binaries for macOS, Linux and Windows on tag push and publishes them as GitHub Releases with checksums; an install script downloads the right binary into ~/.local/bin; prep update replaces the binary from the latest release, or names the right command when prep was installed by Homebrew or go install; go install keeps working. The Claude Code integration ships as a plugin (01M46P009GD5Y6ACBYRMN00D4V) that checks the binary version. Implemented in 01M46ARP808AXT8MABAYMNHSY3. Alternatives: go install only (needs a Go toolchain, no updates); Homebrew first (macOS and Linux only).
