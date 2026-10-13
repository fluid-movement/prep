---
title: Human-led completion contracts and TUI-first workflow
kind: manual
tags:
  - workflow
  - tui
priority: high
---

Implement the direction agreed in the product interview: prep is a contract interface for a human-led, mostly linear development process, not an implementation or commit orchestrator. The agent executes project checks; prep validates and records evidence. Require an explicit project completion policy before any issue becomes ready. Separate implementation completion from versioned human approval, and support explicit reopening with preserved history. Focus the human interface on a TUI Inbox and completion-review bundle; migrate useful desktop behavior and tests before deleting the desktop app. Keep CLI-first agent integration, lean knowledge in OK Format, and explicit cross-project uncertainty. First deliver the completion/review contract plus TUI review and Inbox; then dependency uncertainty and knowledge experiments. Harden local writes, but defer distributed coordination, MCP, dedicated external adapters, and small-change shortcuts. Evaluation initially prioritizes reliable contract compliance, with focused repeated comparisons and deterministic CI gates. This issue records agreed future requirements, not claims about current behavior. Related existing work: 01M4JVP89F3PGDHYETHEPS3YBB (desktop experiment), 01M485SARRF3924985Z0Q4E6SS (knowledge review), 01M4G30VB0SNZT4AJ16XR0FD88 (CI gaps), 01M4J8JW0Z4746FVC99577A6QX (knowledge freshness), 01M4QKTQVR3J6PJZAJM0MVWQR0 (eval framework), 01M46ARR6G4K7NZBYZQJZGBCR0 (branch claims), and 01M46ARY20YT041R18NNMPY95K (deferred MCP).

## Open questions
