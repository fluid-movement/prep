---
type: feature
title: Agent context
description: 'How agents get context without loading everything: prime at session start, guide for the chosen issue, then the files it points to; knowledge retrieval order.'
status: stable
generated:
  by: claude-code/2.1.289
  at: 2026-10-06T08:37:30Z
scope:
  - internal/domain/guide.go
  - internal/cli/read.go
confirmed_commit: 33abc6e396dee255da96c8fc22e75aa4119fa606
---

# Agent context

Applies when changing what `prep prime` or `prep guide` show, or how knowledge is retrieved.

- **Context layering.** `prep prime` runs at session start (the [Claude Code integration](/components/claude-code.md) plugin hook, with `--hook` so it is silent outside prep projects) and gives a capped briefing of pointers: the bootstrap alert, top-level parents with progress, actionable issues, stale issues, claims, check failures. The agent then runs `prep guide <id>` for the issue it works on, which names the step, unmet gates, the files to read and the commands that write outputs. Only then does it open files. Each layer points to the next, so nothing large enters the context unless the work needs it.
- **Knowledge retrieval** in `prep guide` lists candidate entries by title and description only, in this order (`Tree.Candidates`): entries linked from the issue's `context.md`, entries linked from the parent chain's context, entries whose `scope` overlaps the files the issue touches, and finally the overview. The agent opens what it needs and follows links from there; OKF `index.md` files list everything.
- **Priority** orders what agents see first: `prep next` and prime's actionable and stale lists sort by priority, then ID, and mark non-medium levels (`!crit`, `!high`, `low`).
- **Guided one-time work.** Work that needs the user's judgment and project-specific sources is planned as a guided issue rather than built in: the knowledge bootstrap (`prep knowledge bootstrap`) and the import of existing work items (`prep import`: one research issue tagged `import` whose context walks through asking for sources, writing candidates as findings for review, skipping candidates whose `Source:` line `prep list --text` finds, and creating open issues with `prep new` ending in `Source: <ref>`; nothing is kept in sync). Both show up through `prep prime` and `prep guide` like any issue.
- **Skill.** The skill is a short pointer to `prep prime` and `prep guide`; the instructions come from the binary, so they always match its version. A static copy (`.claude/skills/prep/SKILL.md`) serves sessions without the binary.
- Read and write commands stay strictly separate, so harness permission rules can allow reads freely and ask before writes ([Design principles](/decisions/design-principles.md)).

Lifecycle steps themselves: [Issue lifecycle](/features/lifecycle.md).
