---
type: feature
title: Knowledge in the issue lifecycle
description: Documentation decisions and knowledge updates during issue completion
generated:
  by: codex
  at: 2026-10-14T12:48:05Z
---

# Knowledge in the issue lifecycle

Use `prep knowledge find|show` to inspect relevant records and legacy descriptive entries. `prep decide [<issue>] --title T --description D` and `prep finding [<issue>] --title T --description D` create immutable knowledge records, optionally linked to an issue and scoped to code paths. New decisions may supersede older records across the knowledge base; status is derived. Show and guide list an issue's linked records by title.

Decision issues require a linked decision record with outcome to complete, research issues a linked finding record. Completion has no documentation flags or gate, and knowledge has no bootstrap or drift-confirmation workflow. The project Definition of Done remains guidance. See [Completion requirements](/features/completion-policy.md) and [Storage format](/conventions/storage-format.md).
