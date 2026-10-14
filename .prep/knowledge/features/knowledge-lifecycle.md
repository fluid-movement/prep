---
type: feature
title: Knowledge review in the issue lifecycle
description: Bounded knowledge guidance and explicit, scoped reviews required for policy-bound completion.
generated:
  by: codex
  at: 2026-10-14T12:48:05Z
---

Policy-bound issues get up to five relevant knowledge candidates in their guide; agents use `prep knowledge find` and `show` to inspect entries. Before completion, every `--docs` entry needs a current explicit `prep knowledge review record <id> --entry <entry> --kind updated|unchanged --observation <text> --scope <work-path>` report. Alternatively, `--no-impact --reason <reason>` requires a matching `--no-impact --reason <reason>` review report. The domain gate is shared by CLI and TUI; the TUI collects the documentation decision before submission.

Reports are attributed agent assessments, not authenticated approvals or proof of semantic correctness. They bind to the ready contract, selected entry content revision (when named), declared work/configuration scopes and a worktree fingerprint. Changed scope, contract or entry content requires a fresh report. Resolution stores only reports used for that completion. The guide shows a bounded candidate list and targeted retrieval commands. Legacy unbound completions remain compatible and do not acquire invented reports. See [Completion policy](/features/completion-policy.md) and [Storage format](/conventions/storage-format.md).
