---
title: Complete a code issue in the same commit as its code
kind: code
parent: 01M48KB1NRFQ1A3VWB9SHDM3TM
tags:
  - workflow
---

`prep complete <id> --commit <hash>` needs the code commit to exist first, so the completion records (resolution, confirmed knowledge) always land in a second commit. Completing a code issue without `--commit` records that its evidence is the commit that adds its `resolution.md`, so the agent completes first and commits code and records together.

- `prep show`, `prep check` and everything that reads evidence resolve it through git (`git log` on the resolution file) once that commit exists; until then the evidence is pending and shown as such.
- `--commit <hash>` keeps working for work committed earlier (other branches, squash merges).
- Evidence stays informational, as decided in 01M46ARQ78P8M40K4FSPVNK531.
- `prep guide` shows `--commit` as optional for code issues, and the skill describes the one-commit flow.

## Open questions
