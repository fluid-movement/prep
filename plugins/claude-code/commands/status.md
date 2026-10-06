---
description: Show every prep issue with its state
allowed-tools: Bash(prep:*), Bash(echo:*)
---

## Briefing

!`prep prime 2>&1 || echo "PREP_MISSING"`

## Issues

!`prep list 2>&1 || echo "PREP_MISSING"`

Present this to the user as the current state of the project: one short line with the totals and any alerts from the briefing, then the issues grouped by state (in progress, ready, defined, open, then done and dropped), each as ID, kind, title and any blocked or progress note. Do not run any write command.

If the output says PREP_MISSING or the command was not found, tell the user the prep binary is not installed (curl -fsSL https://raw.githubusercontent.com/fluid-movement/prep/main/install.sh | sh) and that the prep skill describes working without it.
