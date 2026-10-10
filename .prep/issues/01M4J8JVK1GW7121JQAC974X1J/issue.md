---
title: 'Targeted knowledge edits: replace text in an entry without rewriting its body'
kind: code
parent: 01M498VPZAAQJKHXGZZ3PWJH2F
tags:
  - knowledge
  - cli
priority: high
---

`prep knowledge update` only replaces an entry's whole body, so every session that changes a few sentences writes its own script (a `kedit.py` was reinvented in at least five sessions, 2026-10-06 to 2026-10-09). The script reads the raw file under `.prep/knowledge`, bypassing prep's reads, applies old/new replacements, and pipes the full body back through `--body-file -`. The ledger analysis (01M4J8780KZRKWHJR4NN8AD85M) counts 25 scripted knowledge writes while `update` existed, and 11 since knowledge reads went through prep.

Add a targeted edit to `prep knowledge update` (or a sibling command): replace an exact text that must occur once in the body, and replace one section by its anchor, so an agent changes an entry without reading or resending all of it. Mismatches fail with the reason, like the Edit tool's.

## Open questions
