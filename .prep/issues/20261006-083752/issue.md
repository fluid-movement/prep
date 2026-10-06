---
title: Collision-free issue IDs (ULID)
kind: code
parent: 20261006-112434
tags:
  - storage
---

Issue IDs are second-resolution timestamps, so two branches can create issues with the same ID; when they merge, both issues' files land in the same directory and git either reports conflicts or silently mixes two issues. Instead of detecting that after the fact, prep new creates IDs that cannot collide: ULIDs (a 48-bit millisecond timestamp and 80 random bits in Crockford base32, 26 characters, sortable by creation time). Everything that takes or shows an ID (directory names, parent and dependency references, CLI arguments, JSON output, the TUI, prep check) accepts them, and issue order by ID keeps meaning creation order.

## Open questions

- Existing issues: keep their timestamp IDs valid next to ULIDs (no churn, old references in commits and notes keep working; ordering compares creation times rather than raw strings), or migrate every issue to a ULID once (one format, but every directory, reference and past commit message changes)? Recommendation: keep them valid.
- How is a ULID shown when space is short (TUI rows, footers, commit subjects)? Today the TUI shows the time part (134231); a ULID's leading characters are its time and repeat for issues created close together, so a short form would take its last characters (for example the last 6). Recommendation: the last 6 characters, and every command also accepts any unique suffix of an ID.
