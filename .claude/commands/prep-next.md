---
description: Show the prep issues that can be worked on now
allowed-tools: Bash(prep:*), Bash(echo:*)
---

!`prep next 2>&1 || echo "PREP_MISSING"`

Present the actionable issues to the user as a short list of ID, kind and title, and suggest which one to take with a one-line reason. If there are none, say so and run nothing else; issues become actionable once they are defined and ready, and `/prep-status` shows where each one stands. Do not run any write command.

If the output says PREP_MISSING or the command was not found, tell the user the prep binary is not installed (`just install` from a prep checkout) and that `.claude/skills/prep/SKILL.md` describes working without it.
