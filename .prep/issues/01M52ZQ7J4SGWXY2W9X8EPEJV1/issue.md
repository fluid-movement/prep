---
title: Delete the desktop app
kind: code
parent: 01M52ZNHXFXHNDATT9WWQ3C87E
tags:
  - desktop
priority: high
---

Remove the macOS desktop app and everything that exists only for it: desktop/ (35 tracked files, about 6,600 lines of Swift), the desktop job in CI, the desktop recipes in the justfile, its knowledge entry /components/desktop.md, and its mentions in other knowledge entries and one Go comment. Done issues about the desktop stay as history. The rethink decided TUI only (01M4ZCMBPJ5FM06F9H70XMPE9H); the failing desktop job is why GitHub CI on main is red since 2026-10-13. JSON output fields the desktop may have been the only reader of are left to the prune sweep (01M52ZS90G529Q80FGM51RR5DS).

## Open questions
