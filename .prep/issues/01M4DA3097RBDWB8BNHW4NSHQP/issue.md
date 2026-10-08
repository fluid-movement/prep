---
title: 'Skill: wide surveys go to subagents that read narrowly'
kind: code
tags:
  - tokens
  - integration
---

Every read in the main context is paid again, as cache reads, on every later request; a survey of many files (a knowledge-base bootstrap, an import, a triage of a whole codebase) is where that adds up most. The ledger of a fathom bootstrap (session 704a70a0) shows subagents keep the main context small (105k at the end against about 187k tokens read), but also that they ran on the parent's model, re-read the same knowledge entries in every agent, and read oversized command output back whole.

The prep skill (both copies: `internal/cli/skill.md` and the plugin's) gains guidance:
- Read knowledge through `prep knowledge list|find|show`, narrowest first; never `cat` entries or whole directories of them.
- Wide surveys go to subagents (one per area); their prompts carry the knowledge-read rule, name the entries or sections each needs, and ask for a short report. Pick a smaller model for survey subagents when the harness allows choosing one.
- Read late and narrow: search and read ranges rather than whole files, and keep command output bounded instead of reading back truncated output files.

## Open questions
