---
title: Delete the desktop app
kind: code
parent: 01M52ZNHXFXHNDATT9WWQ3C87E
tags:
  - desktop
priority: high
---

Remove desktop/ (the macOS Swift app), its CI job, justfile recipes, knowledge entries and every reference to it. Its Inbox and review were never migrated because the Inbox itself is removed. The failing desktop job is why GitHub CI on main is red since 2026-10-13. Decided in 01M4ZCMBPJ5FM06F9H70XMPE9H; replaces 01M4TSY329F505HDC22HFYPD42. Delete the code, its tests, fixtures, eval cases, record handling and knowledge text together; nothing stays behind for compatibility. Done when nothing references the removed feature and GitHub CI on main is green.

## Open questions
