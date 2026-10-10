---
type: component
title: Project registry
description: 'internal/projects: the machine-local list of prep projects (root, name, last used) that commands record and prep projects and the desktop app read.'
generated:
  by: claude-code/opus-5.5
  at: 2026-10-10T12:18:55Z
scope:
  - internal/projects/**
---

# Project registry

Applies when changing how prep remembers projects or how clients switch between them.

- **Where**: `$XDG_STATE_HOME/prep/projects.json`, default `~/.local/state/prep/projects.json`. It is machine-local state, not a choice, so it is not in the user configuration ([Harness setup and user configuration](/components/setup.md)). A damaged file is treated as empty and rebuilt from the next touches.
- **Recording**: `cli.Main` calls `recordProject` after every successful command that opened a project; it runs `projects.Touch(root, now)`, which writes at most once a minute per root, so a burst of commands costs one write. A failure never fails the command. Concurrent commands may lose a touch; the next one repairs it.
- **Listing**: `projects.List` returns projects whose `.prep` still exists, newest first, and rewrites the file when it dropped missing ones. `prep projects` prints name, last use and root; `--json` gives `{"projects": [{"root", "name", "last_used"}]}`. It needs no project.
- **Clients**: the desktop app's project switcher lists it.
- **Tests**: CLI tests point `XDG_STATE_HOME` at a temporary directory in `TestMain`, so commands they run never record into the real registry.
