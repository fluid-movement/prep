---
title: One lock per write command; state file written last
kind: code
parent: 01M52ZNHXFXHNDATT9WWQ3C87E
depends_on:
  - 01M52ZQAGVQ5HX0GTBX7R2P47G
tags:
  - storage
---

Hold the checkout lock for the whole write command (load, plan, write) and drop the hash-snapshot stale-plan check for the CLI; the TUI re-reads under the lock before writing. Remove the transaction journal and roll-forward recovery. Write the state-defining file last (for example resolution.md after its supporting records). Decided in 01M4ZCMFJ5P67XNSMKX1NPANDF. Delete the code, its tests, fixtures, eval cases, record handling and knowledge text together; nothing stays behind for compatibility. Done when nothing references the removed feature and GitHub CI on main is green.

## Open questions
