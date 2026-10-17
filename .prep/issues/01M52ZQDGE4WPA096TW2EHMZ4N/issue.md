---
title: Remove review approvals and the Inbox; keep a minimal reopen
kind: code
parent: 01M52ZNHXFXHNDATT9WWQ3C87E
tags:
  - workflow
  - tui
priority: high
---

Remove human approval records and the Inbox, per the review decision (01M4ZCM3ZJDTM3JRHS2N1NMF5H): prep review show and approve, resolution reviews (resolution-reviews.md), resolution generations and resolution-history.md, approval prerequisites on dependencies (depends_on_approval), prep dependency override with dependency-overrides.md, the TUI Inbox screen, the completion bundle and its review snapshot. Reopen becomes minimal: it needs a reason, appends the old resolution to history.md, removes resolution.md and returns the issue to ready. This repository's records are converted once, with no compatibility code. The completion contract machinery (policy, evidence, exceptions, knowledge review reports) stays until 01M52ZQAGVQ5HX0GTBX7R2P47G. Delete the code, its tests, fixtures, eval cases and knowledge text together. Done when nothing references the removed parts and GitHub CI on main is green.

## Open questions
