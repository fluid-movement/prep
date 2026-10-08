---
type: component
title: Harness setup and user configuration
description: prep setup installs, refreshes and removes harness integrations through the setup.Harness interface; ~/.config/prep/config.yaml records which harnesses the user chose.
status: stable
generated:
  by: claude-code/2.1.289
  at: 2026-10-06T09:05:34Z
scope:
  - internal/setup
  - internal/userconfig
  - internal/cli/setup.go
  - internal/tui/checklist.go
confirmed_commit: 1a82fd30cc1dd585eea3d2e213d7ed86dffc0ed2
---

# Harness setup and user configuration

Applies when adding a harness integration or changing how prep setup or the user configuration work.

- **Harness interface** (`internal/setup`): `Name`, `Title`, `Detect` (present on this machine), `Installed` (integration version, if any), `Install(version)` (also updates; returns warnings next to the error, carried in `Result.Warnings` and printed by `prep setup` as `warning:` lines), `Remove`. Implementations register themselves with `setup.Register` from `init`. `Apply(selected, version)` installs or updates every selected harness and removes every other installed one; `Refresh(chosen, version)` brings the chosen ones to the binary's version and skips harnesses no longer on the machine; one failing harness never blocks the others. `Survey` reports detected, installed and chosen per harness.
- **User configuration** (`internal/userconfig`): `$XDG_CONFIG_HOME/prep/config.yaml`, default `~/.config/prep/config.yaml`, strict YAML, atomic writes. It records only choices (`harnesses`, and `tui.mouse` for [TUI mouse](/components/tui-mouse.md) capture, unset meaning on);
 installed versions are always asked from the harness ([Design principles](/decisions/design-principles.md), "a small database"). `tui.layout` (`single` for one pane at any width, toggled with `z` in the TUI) is the other TUI choice. It is the place for later per-user choices such as a TUI theme.
- **prep setup** (`internal/cli/setup.go`): without flags it asks with a design-system checklist (`tui.RunChecklist`) that preselects chosen, installed and detected harnesses and reads keys from `/dev/tty` (`CONIN$` on Windows), so it works under `curl | sh`; cancelling changes nothing. `--harness a,b` selects exactly those, `--remove name` removes one, `--refresh` updates the chosen ones; all support `--json` and report one result per harness (installed, updated, up to date, removed, skipped, failed). Without a terminal it names the non-interactive command. With no harness registered it says there is nothing to do.
- **Callers**: `install.sh` runs `prep setup` on the terminal after installing the binary, and `prep update` runs the new binary's `prep setup --refresh` ([Release, install and update](/components/release.md)).
- The first harness is Claude Code, registered by `internal/setup/claudecode` and imported by the CLI; see [Claude Code integration](/components/claude-code.md). Tests use fake harnesses registered in the test.
