---
type: component
title: Project registry
description: 'The machine-local registry of prep projects: where it lives, how commands record into it, project identity (ID, name, remote) and how a project is resolved.'
generated:
  by: claude-code/opus-5.5
  at: 2026-10-10T12:18:55Z
scope:
  - internal/projects/**
---

# Project registry

Applies when changing how prep remembers projects, how a project is identified, or how clients switch between and resolve projects.

- **Identity**: a project is its permanent ID (`id:` in [project.md](/conventions/storage-format.md), a ULID shared by every clone). The `name:` is a handle (declared, else `DefaultName`: origin's last path segment, else the directory name, through `SanitizeName`); the remote (`gitx.Remote` through `NormalizeRemote`: `host/owner/repo`, so SSH, scp-like and HTTPS forms are equal; local paths stay cleaned paths) only says where to clone it from. Renaming, changing the remote or moving the directory keeps the identity; a fork that keeps project.md is a clone of the same project until `prep init --new-id`. Decided in 01M4N5ARVE6J5APZYJ7VA0N464 (D2).
- **Where**: `$XDG_STATE_HOME/prep/projects.json`, default `~/.local/state/prep/projects.json`. It is machine-local state, not a choice, so it is not in the user configuration ([Harness setup and user configuration](/components/setup.md)). A damaged file is treated as empty and rebuilt from the next touches.
- **Rows**: `root`, `id`, `name`, `remote`, `last_used`. `Describe(root)` reads the identity (project.md through `mdstore.ReadProject`, remote from git); it never fails, leaving what it cannot read empty.
- **Recording**: `cli.Main` calls `recordProject` after every successful command that opened a project; it runs `projects.Touch(root, now)`, which writes at most once a minute per root, so a burst of commands costs one write. A failure never fails the command. Concurrent commands may lose a touch; the next one repairs it. `Add(dir, at)` registers a checkout prep has not run in (it must have `.prep/project.md`); `Remove(query)` forgets a root path, or every clone of the project a name or ID resolves to. Neither touches the project.
- **Listing**: `List` returns projects whose `.prep` still exists, newest first, with the identity re-read from each checkout, and rewrites the file when it dropped missing ones or an identity changed. `Collision` (derived, never stored) marks a name that projects with different IDs share. Clones (one ID, several roots) are separate rows. `prep projects` prints them; `--json` gives `{"projects": [{"root", "id", "name", "remote", "last_used", "collision"}]}`. It needs no project.
- **Resolution**: `Resolve(query)` tries an exact name, then an ID or a unique suffix of one (at least 4 characters, any case), then a normalized remote, then a path to a checkout (registered or not). Matches are grouped by project (ID, else root), so clones resolve to the most recently used one; matches from several projects are `E_AMBIGUOUS` listing them, no match is `E_NOT_FOUND` listing the registered names. `ByID(id)` returns the newest checkout of a project.
- **Commands**: `prep projects [add <dir> | remove <q>]`; `Main` records into the registry. `prep init` writes the identity (`DefaultName`), `prep init --new-id` replaces only the ID, `prep fix` writes a missing ID and name (`cli.fixIdentity`).
- **Clients**: the [desktop app](/components/desktop.md)'s project switcher lists it.
- **Tests**: CLI tests point `XDG_STATE_HOME` at a temporary directory in `TestMain`, so commands they run never record into the real registry; store tests pass a fixed identity to `Init`, and the contract fixtures carry `id: 01M4AAAAAAAAAAAAAAAAAAAAAA`.
