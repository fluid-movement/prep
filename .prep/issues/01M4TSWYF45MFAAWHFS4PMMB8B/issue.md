---
title: Local write safety without distributed orchestration
kind: code
parent: 01M4TST3C4ZSZDKA5WKBA3ND4H
depends_on:
  - 01M4TSWX8HF105ZQQKS50MH6RP
tags:
  - workflow
---

Harden concurrent writes within a checkout, investigate check-then-rename races and interrupted multi-file changes, and cover guarantees with deterministic conflict/recovery tests. Clearly document branch and worktree claim visibility limitations. Keep prep human-led and mostly linear; defer distributed coordination and scheduling. Occasional orchestrator-assigned subagents on ready issues do not justify a remote reservation service now. Preserve Markdown records and domain/storage boundaries. Related historical work: 01M4E7XR44YR718HVTPD5N1BWT and 01M46ARR6G4K7NZBYZQJZGBCR0.

## Open questions
