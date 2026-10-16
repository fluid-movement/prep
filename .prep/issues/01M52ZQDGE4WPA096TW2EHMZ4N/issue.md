---
title: Remove review approvals and the Inbox; keep a minimal reopen
kind: code
parent: 01M52ZNHXFXHNDATT9WWQ3C87E
depends_on:
  - 01M52ZQAGVQ5HX0GTBX7R2P47G
tags:
  - workflow
  - tui
priority: high
---

Remove prep review, resolution reviews and generations, approval prerequisites on dependencies (depends_on_approval), dependency override, approvals.md, the TUI Inbox screen and completion bundle. Reopen stays minimal: it moves the old resolution into history.md and returns the issue to ready. Decided in 01M4ZCM3ZJDTM3JRHS2N1NMF5H. Delete the code, its tests, fixtures, eval cases, record handling and knowledge text together; nothing stays behind for compatibility. Done when nothing references the removed feature and GitHub CI on main is green.

## Open questions
