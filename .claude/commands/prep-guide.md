---
description: Show the prep guide for an issue and work from it
argument-hint: <issue-id>
allowed-tools: Bash(prep:*), Bash(echo:*)
---

!`prep guide $ARGUMENTS 2>&1 || echo "PREP_MISSING_OR_FAILED"`

This is the guide for issue $ARGUMENTS. Summarize for the user the issue's state, its current step and the unmet gates in a few lines, then work on that step as the guide says, following the prep skill. Before any transition command (`define`, `ready`, `claim`, `complete`, `drop`), confirm with the user when the guide says the step is theirs to settle.

If no ID was given, ask for one and point at `/prep-status`. If the binary is missing, tell the user it is not installed (`just install` from a prep checkout) and that `.claude/skills/prep/SKILL.md` describes working without it.
