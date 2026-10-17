---
type: decision
title: "Direction: Markdown in git as truth, a derived local SQLite index as the read source"
description: "Direction: Markdown in git as truth, a derived local SQLite index as the read source"
date: "2026-10-16"
issue: 01M4ZCMFJ5P67XNSMKX1NPANDF
---

User direction with agent recommendation, 2026-10-16. Markdown stays the committed source of truth because it merges; SQLite cannot be resolved in a merge conflict. One file per record named by ULID, so new records never conflict; old records are never changed (a new decision states what it supersedes; status is derived). A gitignored SQLite index, covering issues and decisions with full-text search, becomes the read source. Prep keeps it fresh itself on every command (stat record files, re-parse changed ones, update on its own writes), not the TUI, because agents run prep without a TUI. Data flows one way, Markdown to index; the index is disposable and rebuilt on demand. Similar to Beads (JSONL in git plus local SQLite), but file-first, so nothing syncs back. Design the record format for it now; build it when measured read time warrants it (about 0.3 s once the contract machinery is gone; decisions may grow by thousands a year).
