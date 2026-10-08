---
title: 'Claude Code: open and close the pane with /prep'
kind: code
parent: 01M498VPZAAQJKHXGZZ3PWJH2F
tags:
  - integration
---

The pane command in Claude Code becomes `/prep [live|project|usage]`, the same command as in Pi (D7 on 01M46ARM9G3492V298TM82KD45): without an argument it toggles the pane, with a view it opens the pane on that view. `/prep:pane` goes away; there are no outside users to keep it for. The command is answered by the panel's hooks without reaching the model, as today, and the [Claude Code side panel](/components/claude-code-panel.md) entry, the plugin's commands and the docs that name `/prep:pane` (TOKEN-EFFICIENCY.md, the token-ledger README) change with it.

## Open questions

- In Claude Code, skills are commands too: the project skill copy `.claude/skills/prep` (kept for cloud sessions) is invoked as `/prep`, and the plugin's skill as `/prep:prep`. Which one gives way, and can a command registered by the hooks module take a bare name next to a skill of the same name?
- Pi keeps only the pane command (D7). Do `/prep:status`, `/prep:next`, `/prep:guide` and `/prep:focus` stay in Claude Code, or go the same way?
