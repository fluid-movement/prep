---
type: decision
title: Plugin refresh must keep sibling plugins
description: Moving the pin uninstalls every plugin of the marketplace, so refresh reinstalls the others, restores the old marketplace on failure and warns. (merges codex.md)
date: "2026-10-17"
scope:
  - internal/setup/claudecode/**
  - internal/setup/codex/**
---

Moving the pin uninstalls every plugin of the marketplace, so refresh reinstalls the others, restores the old marketplace on failure and warns. (merges codex.md)

Removing the marketplace uninstalls all its plugins (stated).
