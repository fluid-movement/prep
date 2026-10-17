---
type: convention
title: Knowledge base conventions
description: How knowledge entries are written, typed, linked, scoped and kept current; precedence when sources disagree.
status: stable
generated:
  by: claude-code/2.1.289
  at: 2026-10-05T00:00:00Z
---

# Knowledge base conventions

Knowledge starts empty and grows from immutable decisions and findings. Existing descriptive entries remain readable through `prep knowledge list|find|show`; they have no creation, update, confirmation or bootstrap command.

- `prep decide [<issue>] --title T --description D [--body text | --body-file path|-] [--scope glob]... [--supersedes record]... [--outcome]` creates a decision; `prep finding [<issue>] --title T --description D [--body ...] [--scope glob]...` creates a finding.
- Each record is one OKF v0.2 file, `knowledge/decisions/<ulid>.md` or `knowledge/findings/<ulid>.md`. Required fields are type, title, a one-line description and date. Issue, scopes, supersedes IDs and decision outcome are optional. A record without scopes is project-wide. Body contains context, choice and why, or the finding.
- Records are never edited. A newer record lists the older record ID in supersedes; status is derived and shown in list, find and show. References accept record IDs or unique suffixes across the bundle.
- Keep rationale, constraints and useful navigation small; lint warns above 8 KiB. Use targeted find/show retrieval. Links are bundle-relative paths and must resolve. Old descriptive entries retain their readable paths.
- Index files are generated, with OKF version 0.2 on the root index. Init configures union merges for indexes; fmt and fix regenerate them after merges. Do not edit indexes.
- Decision completion requires a linked decision record with outcome; research completion requires a linked finding. Completion has no documentation flags or gate. Code drift and confirmation markers are no longer computed or exposed.
