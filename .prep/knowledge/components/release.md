---
type: component
title: Release, install and update
description: How prep is released (GoReleaser on v* tags), installed (install.sh, go install, just install) and updated (prep update with checksum verification).
status: stable
generated:
  by: claude-code/2.1.289
  at: 2026-10-06T08:42:41Z
scope:
  - .goreleaser.yaml
  - .github/workflows/release.yml
  - install.sh
  - internal/update
  - internal/cli/update.go
confirmed_commit: f99448c473031bdb5855c29c9d5116ab252f5572
---

# Release, install and update

Applies when changing how prep is built for distribution, installed or updated. Decided in 01M46ARN8RW78DATMQV6P9NVYK.

- **Release**: pushing a `v*` tag runs `.github/workflows/release.yml`, which tests and then runs GoReleaser (`.goreleaser.yaml`): darwin, linux and windows on amd64 and arm64, `CGO_ENABLED=0`, `-trimpath`, version stamped into `internal/cli.Version` from the tag. Archives are `prep_<version>_<os>_<arch>.tar.gz` (zip on Windows) plus `checksums.txt` (sha256). The archive name is a contract with `update.ArchiveName` and `install.sh`.
- **Preparing a release**: `just release vX.Y.Z` sets the version in the Claude Code plugin manifests, runs the tests, commits and creates an annotated tag; pushing the tag (`git push origin main vX.Y.Z`) publishes. The release workflow first checks that both manifests carry the tag's version.
- **Versions**: release builds report the tag; `just build`/`just install` stamp `git describe`; `just install` then runs `prep setup --refresh` with `PREP_PLUGIN_SOURCE` set to the checkout, so a developer's Claude Code plugin follows the code (a failed refresh only prints a note); `go install` builds fall back to the module version from the build info; anything else is `dev`. `update.IsRelease` decides whether a version is a release; development builds are not: Go pseudo-versions of untagged builds, `git describe` versions after a tag (`v0.1.0-14-gaa2b6fc`, optionally `-dirty`) and `-dirty` builds.
- **install.sh** (POSIX sh, curl or wget): resolves the latest tag from the releases redirect or uses `PREP_VERSION`, downloads the archive and `checksums.txt`, verifies with `sha256sum` or `shasum`, and installs into `PREP_BINDIR` or `~/.local/bin` (the XDG location for user executables, no root needed; most Linux distributions put it on PATH once it exists, macOS never does). When the directory is not on PATH it prints a ready-to-paste line for the user's shell: `~/.zshrc`, `~/.bashrc` (`~/.bash_profile` on macOS), `fish_add_path`, else `~/.profile`. It never edits shell files itself. `PREP_RELEASE_URL` points it at another release host (used for testing).
- After installing, install.sh runs `prep setup` on the terminal (`</dev/tty`) when one is attached, else prints the command; prep update runs the new binary's `prep setup --refresh` after replacing itself, so harness integrations follow the binary ([Harness setup](/components/setup.md)).
- **prep update** (`internal/update`, standard library only): reads `/repos/fluid-movement/prep/releases/latest`, compares versions (`Newer`), downloads the archive for `runtime.GOOS/GOARCH` and `checksums.txt`, refuses on a mismatch, extracts the binary and renames it over `os.Executable()` (on Windows the old binary is moved aside to `.old` first and moved back if the new one cannot take its place). `--check` only reports. Development builds and installs managed by Homebrew or `go install` (`update.Managed`) get the command to use instead.
- **Verification**: `goreleaser check` validates the configuration, and a snapshot build (`goreleaser release --snapshot --clean --skip=publish`, output in the ignored `dist/`) produces the archive names above. Tests run update against an httptest release server (tar.gz and zip, checksum mismatch, missing platform); install.sh was exercised against a local fake release and checked with dash.
