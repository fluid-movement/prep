---
title: Convert this repository's knowledge to decision records
kind: code
parent: 01M52ZNHXFXHNDATT9WWQ3C87E
depends_on:
  - 01M52ZRB3VY0MSZ0HZWJ2SEC24
tags:
  - knowledge
---

Turn this repository's knowledge into decision and finding records, once the knowledge model (01M52ZRB3VY0MSZ0HZWJ2SEC24) exists. Decided in 01M4ZCM5XQ3J6TM1N4AZ330ZH7 and 01M4ZCK8C17RDGBVCTMQPGMEK7; working method agreed with the user on 2026-10-17.

- The 162 issue decisions (in 80 issues) convert mechanically, without review: each becomes a decision record linked to its issue, with its title, date and body; the 8 that supersede an earlier one map --supersedes D<n> to the record ID.
- The 43 descriptive entries are drafted by a subagent as one table per area (components, TUI, features and conventions): each statement that is a decision or constraint becomes a record (title, one-line description, why, code scopes from the entry); everything else is marked delete. The user reviews each table in chat in batches of about ten entries; only approved rows are applied. Deleted prose is not kept anywhere (git history aside); nothing is reworded to survive.
- The project overview moves to a new AGENTS.md (what prep is, where things live), which Codex loads, plus a CLAUDE.md that imports it for Claude Code. The README stays user-facing.
- Rewrite design principle 9 as 'issues are the work; knowledge is what was decided and learned'.
- Other prep projects (fathom, fpa-api, tuner, fpa-profiles) are out of scope; their entries stay readable, and each gets its own conversion issue when worked on.

Drafting the tables can start before the knowledge model lands; applying them waits for it.

## Open questions
